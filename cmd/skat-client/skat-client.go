package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/horazont/webskat/internal/frontend/singleuser"
	"github.com/horazont/webskat/internal/skat"
)

var (
	serverAddress  = flag.String("client.server-address", "127.0.0.1:5023", "")
	serverPassword = flag.String("client.server-password", "foobar2342", "")
	enableColor    = true
)

func SimpleTimeout(f func(ctx context.Context) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return f(ctx)
}

func backsideColor() string {
	if !enableColor {
		return ""
	}
	return "\x1b[30;47m"
}

func color(s skat.Suit) string {
	if !enableColor {
		return ""
	}

	if s == skat.SuitSpades || s == skat.SuitClubs {
		return "\x1b[30;47m"
	} else {
		return "\x1b[31;47m"
	}
}

func resetColor() string {
	if !enableColor {
		return ""
	}

	return "\x1b[0m"
}

func ppCardLine1(c skat.Card) string {
	return fmt.Sprintf("%s%-2s%s", color(c.Suit), c.Type.Pretty(), resetColor())
}

func ppCardLine2(c skat.Card) string {
	return fmt.Sprintf("%s %s%s", color(c.Suit), c.Suit.Pretty(), resetColor())
}

func ppBlindCardLine1() string {
	return fmt.Sprintf("%s▗▖%s", backsideColor(), resetColor())
}

func ppBlindCardLine2() string {
	return fmt.Sprintf("%s▝▘%s", backsideColor(), resetColor())
}

func renderCardRow(cards skat.CardSet, withNumbers bool) {
	for _, card := range cards {
		fmt.Printf(" %s", ppCardLine1(card))
	}
	fmt.Printf("\n")
	for _, card := range cards {
		fmt.Printf(" %s", ppCardLine2(card))
	}
	if withNumbers {
		fmt.Printf("\n")
		for i := range cards {
			fmt.Printf(" %-2d", i)
		}
	}
}

func renderBlindedCardRow(ncards int) {
	for i := 0; i < ncards; i = i + 1 {
		fmt.Printf(" %s", ppBlindCardLine1())
	}
	fmt.Printf("\n")
	for i := 0; i < ncards; i = i + 1 {
		fmt.Printf(" %s", ppBlindCardLine2())
	}
}

func startViewEx(header string, hand skat.CardSet, withNumbers bool) {
	fmt.Printf("\n\n=== %s ===\n\n", header)
	if hand != nil {
		fmt.Printf("Your hand:\n")
		renderCardRow(hand, withNumbers)
		fmt.Printf("\n\n")
	}
}

func startView(header string, hand skat.CardSet) {
	startViewEx(header, hand, false)
}

func endView() {
	fmt.Printf("\n\n")
}

func recurringPrompt(prompt string, f func(resp string) error) error {
	for {
		fmt.Printf("%s: ", prompt)
		var resp string
		if _, err := fmt.Scanln(&resp); err != nil {
			return err
		}

		if err := f(resp); err != nil {
			fmt.Printf("invalid input: %s\n", err)
			continue
		}
		return nil
	}
}

func actionChoice(prompt string, actions map[string]string) (string, error) {
	var result string
	err := recurringPrompt(prompt, func(resp string) error {
		var ok bool
		result, ok = actions[resp]
		if !ok {
			return fmt.Errorf("not a valid action: %s", resp)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

func intOrAction(prompt string, actions map[string]string, validateInt func(i int) error) (results string, resulti int, err error) {
	err = recurringPrompt(prompt, func(resp string) (err error) {
		var ok bool
		results, ok = actions[resp]
		if ok {
			return nil
		}

		resulti, err = strconv.Atoi(resp)
		if err != nil {
			return fmt.Errorf("not a valid action or number (%s): %s", err, resp)
		}

		err = validateInt(resulti)
		if err != nil {
			return fmt.Errorf("numeric input not valid: %s", err)
		}

		return nil
	})
	if err != nil {
		return "", 0, err
	}
	return results, resulti, nil
}

func sortOrder(power int, effectiveSuit int) int {
	// we flip the least and second-least bits in effectiveSuit in order
	// to sort red and black cards alternatingly
	effectiveSuit = ((effectiveSuit & 1) << 1) | ((effectiveSuit & 2) >> 1) | (effectiveSuit &^ 3)
	suitOffset := int(effectiveSuit) << 16
	return power + suitOffset
}

func sortHand(gameType skat.GameType, hand skat.CardSet) {
	if gameType != skat.InvalidGameType {
		// sort based on the game type
		sort.SliceStable(
			hand[:],
			func(i, j int) bool {
				// must return true if hand[i] < hand[j]
				iOrder := sortOrder(hand[i].RelativePower(gameType), int(hand[i].EffectiveSuit(gameType)))
				jOrder := sortOrder(hand[j].RelativePower(gameType), int(hand[j].EffectiveSuit(gameType)))
				return iOrder < jOrder
			},
		)
	} else {
		// sort naively
		sort.SliceStable(
			hand[:],
			func(i, j int) bool {
				// must return true if hand[i] < hand[j]
				return sortOrder(int(hand[i].Type), int(hand[i].Suit)) < sortOrder(int(hand[j].Type), int(hand[j].Suit))
			},
		)
	}
}

func sortedHand(gameType skat.GameType, hand skat.CardSet) skat.CardSet {
	result := hand.Copy()
	sortHand(gameType, result)
	return result
}

const (
	DoTakeSkat = "takeSkat"
	DoDeclare  = "declare"
)

const (
	DoDeclareSchneider      = "schneider"
	DoDeclareSchwarz        = "schwarz"
	DoDeclareOuvert         = "ouvert"
	DoDeclareSelectDiamonds = "diamonds"
	DoDeclareSelectHearts   = "hearts"
	DoDeclareSelectSpades   = "spades"
	DoDeclareSelectClubs    = "clubs"
	DoDeclareSelectGrand    = "grand"
	DoDeclareSelectNull     = "null"
	DoDeclareCancel         = "cancel"
	DoDeclareResetPushset   = "resetPushset"
	DoDeclareDone           = "done"
)

var (
	ErrAbortedByUser = errors.New("action aborted by user")
)

func composeGameDeclaration(l *zap.SugaredLogger, gc *singleuser.GameClient, st singleuser.ClientState, hand skat.CardSet) error {
	var pushset skat.CardSet
	var gtype skat.GameType
	var modifiers skat.GameModifier
	handCopy := sortedHand(gtype, hand)
	for {
		modifiers = modifiers.Normalized()

		startViewEx("Compose game declaration", handCopy, true)
		if len(pushset) == 0 {
			fmt.Printf("No cards to push\n")
		} else {
			fmt.Printf("Pushing:\n")
			renderCardRow(pushset, false)
			fmt.Printf("\n")
		}

		if gtype != 0 {
			fmt.Printf("Game type: %s  Modifiers: %s\n", gtype.Pretty(), modifiers.Pretty())
		} else {
			fmt.Printf("No game type selected\n")
		}

		action, cardIndex, err := intOrAction(
			"Declare:\n"+
				" toggle [s]chneider\n"+
				" toggle [S]chwarz\n"+
				" toggle [o]uvert\n"+
				" [d]iamonds "+skat.SuitDiamonds.Pretty()+"\n"+
				" [h]earts "+skat.SuitHearts.Pretty()+"\n"+
				" [c]lubs "+skat.SuitClubs.Pretty()+"\n"+
				" s[p]ades "+skat.SuitSpades.Pretty()+"\n"+
				" [g]rand\n"+
				" [n]ull\n"+
				" [r]eset pushed cards\n"+
				" [0-9] push card\n"+
				" [x] cancel\n"+
				" [y] declare!\n",
			map[string]string{
				"s": DoDeclareSchneider,
				"S": DoDeclareSchwarz,
				"o": DoDeclareOuvert,
				"d": DoDeclareSelectDiamonds,
				"h": DoDeclareSelectHearts,
				"c": DoDeclareSelectClubs,
				"p": DoDeclareSelectSpades,
				"g": DoDeclareSelectGrand,
				"n": DoDeclareSelectNull,
				"r": DoDeclareResetPushset,
				"x": DoDeclareCancel,
				"y": DoDeclareDone,
			},
			func(v int) error {
				if len(handCopy) == 10 {
					return fmt.Errorf("cannot push more cards")
				}
				if v < 0 || v >= len(handCopy) {
					return fmt.Errorf("card index out of range")
				}
				return nil
			},
		)

		if err != nil {
			return err
		}

		switch action {
		case "":
			{
				// push card
				card := handCopy[cardIndex]
				handCopy, _ = handCopy.Pop(card)
				pushset, _ = pushset.Push(card)
			}
		case DoDeclareResetPushset:
			{
				handCopy = sortedHand(gtype, hand)
				pushset = nil
			}
		case DoDeclareSelectDiamonds:
			{
				gtype = skat.GameTypeDiamonds
				sortHand(gtype, handCopy)
			}
		case DoDeclareSelectHearts:
			{
				gtype = skat.GameTypeHearts
				sortHand(gtype, handCopy)
			}
		case DoDeclareSelectSpades:
			{
				gtype = skat.GameTypeSpades
				sortHand(gtype, handCopy)
			}
		case DoDeclareSelectClubs:
			{
				gtype = skat.GameTypeClubs
				sortHand(gtype, handCopy)
			}
		case DoDeclareSelectGrand:
			{
				gtype = skat.GameTypeGrand
				sortHand(gtype, handCopy)
			}
		case DoDeclareSelectNull:
			{
				gtype = skat.GameTypeNull
				sortHand(gtype, handCopy)
			}
		case DoDeclareSchneider:
			{
				if modifiers.Test(skat.GameModifierSchneiderAnnounced) {
					modifiers = modifiers.Without(skat.GameModifierSchneiderAnnounced).Without(skat.GameModifierSchwarzAnnounced)
				} else {
					modifiers = modifiers.With(skat.GameModifierSchneiderAnnounced)
				}
			}
		case DoDeclareSchwarz:
			{
				if modifiers.Test(skat.GameModifierSchwarzAnnounced) {
					modifiers = modifiers.Without(skat.GameModifierSchwarzAnnounced)
				} else {
					modifiers = modifiers.With(skat.GameModifierSchwarzAnnounced)
				}
			}
		case DoDeclareOuvert:
			{
				if modifiers.Test(skat.GameModifierOuvert) {
					modifiers = modifiers.Without(skat.GameModifierOuvert)
				} else {
					modifiers = modifiers.With(skat.GameModifierOuvert)
				}
			}
		case DoDeclareCancel:
			{
				return ErrAbortedByUser
			}
		case DoDeclareDone:
			{
				if gtype == 0 {
					fmt.Printf("game type required!\n")
				} else if len(handCopy) != 10 {
					fmt.Printf("more cards need to be pushed\n")
				} else {
					err = SimpleTimeout(func(ctx context.Context) error {
						return gc.Declare(ctx, gtype, modifiers, pushset)
					})
					if err != nil {
						fmt.Printf("failed to declare game: %s\n", err)
					} else {
						return nil
					}
				}
			}
		}
	}
}

func declarationPhaseDeclarer(l *zap.SugaredLogger, gc *singleuser.GameClient, st singleuser.ClientState) {
	gs := st.GameState
	startViewEx("Declare a game!", sortedHand(skat.InvalidGameType, gs.Hand), true)

	for {
		action, err := actionChoice("[t]ake Skat or [d]eclare a game", map[string]string{
			"t": DoTakeSkat,
			"d": DoDeclare,
		})
		if err != nil {
			l.Fatalw("failed to read input", "err", err)
		}

		switch action {
		case DoTakeSkat:
			{
				err = SimpleTimeout(func(ctx context.Context) error {
					return gc.TakeSkat(ctx)
				})
				if err != nil {
					fmt.Printf("failed to take skat: %s\n", err)
				} else {
					return
				}
			}
		case DoDeclare:
			{
				err := composeGameDeclaration(l, gc, st, gs.Hand)
				if err == ErrAbortedByUser {
					continue
				}
				if err != nil {
					l.Fatalw("something went wrong", "err", err)
				}
				return
			}
		}
	}
}

func playingPhase(l *zap.SugaredLogger, gc *singleuser.GameClient, st singleuser.ClientState) {
	gs := st.GameState
	myTurn := st.PlayerIndex == gs.CurrentPlayer

	hand := sortedHand(gs.GameType, gs.Hand)

	if myTurn {
		startView("Your turn", nil)
	} else {
		startView("Waiting for other player", nil)
	}

	for i, playerInfo := range gs.Players {
		if i == st.PlayerIndex {
			fmt.Printf("Your hand:\n")
			renderCardRow(hand, myTurn)
		} else {
			fmt.Printf("Player %d:\n", i)
			renderBlindedCardRow(playerInfo.Ncards)
		}
		fmt.Printf("\n")
	}

	fmt.Printf("Table:\n")
	renderCardRow(gs.Table, false)
	fmt.Printf("\n")

	if !myTurn {
		return
	}

	for {
		_, cardIndex, err := intOrAction(
			"pick a card",
			map[string]string{},
			func(v int) error {
				if v < 0 || v >= len(hand) {
					return fmt.Errorf("card number out of bounds")
				}
				return nil
			},
		)

		if err != nil {
			l.Fatalw("input error", "err", err)
		}

		card := hand[cardIndex]
		err = SimpleTimeout(func(ctx context.Context) error {
			return gc.PlayCard(ctx, card)
		})
		if err != nil {
			fmt.Printf("failed to play card: %s\n", err)
		} else {
			return
		}
	}
}

func HandleGameState(l *zap.SugaredLogger, gc *singleuser.GameClient, st singleuser.ClientState) {
	gs := st.GameState
	switch gs.Phase {
	case skat.PhaseInit:
		{
			if !gs.Players[st.PlayerIndex].SeedProvided {
				seed, err := skat.GenerateSeed()
				if err != nil {
					l.Fatalw("failed to generate game seed",
						"err", err,
					)
				}

				err = SimpleTimeout(func(ctx context.Context) error {
					return gc.SetSeed(ctx, seed)
				})
				if err != nil {
					l.Fatalw("failed to set seed",
						"err", err,
					)
				}
				return
			}

			startView("Waiting for other players ...", nil)
			for index, playerInfo := range gs.Players {
				ready_s := "Not ready"
				if playerInfo.SeedProvided {
					ready_s = "Ready"
				}
				fmt.Printf("  %d: %s\n", index+1, ready_s)
			}
			endView()
		}
	case skat.PhaseBidding:
		{
			bs := gs.BiddingState
			startView("Bidding", sortedHand(skat.GameTypeGrand, gs.Hand))
			if bs.LastBid != skat.BidNone {
				fmt.Printf("Current highest: %d\n", bs.LastBid)
			}
			if bs.Caller == st.PlayerIndex && !bs.AwaitingResponse {
				action, call, err := intOrAction("[p]ass or call", map[string]string{
					"p": "pass",
				}, func(v int) error {
					if v < 18 {
						return fmt.Errorf("must be at least 18")
					}
					if v < bs.LastBid {
						return fmt.Errorf("must be higher than last bid")
					}
					return nil
				})
				if err != nil {
					l.Fatalw("bogus input", "err", err)
				}

				if action == "pass" {
					call = skat.BidPass
				}

				err = SimpleTimeout(func(ctx context.Context) error {
					return gc.CallBid(ctx, call)
				})
				if err != nil {
					l.Fatalw("failed to call bid",
						"err", err,
					)
				}
			} else if bs.Responder == st.PlayerIndex {
				if !bs.AwaitingResponse {
					fmt.Printf("You will have to respond\n")
				} else {
					fmt.Printf("Awaiting your response [h/p]:")
					var resp string

					_, err := fmt.Scanln(&resp)
					if err != nil || (resp != "h" && resp != "p") {
						l.Fatalw("bogus input",
							"err", err,
							"resp", resp)
					}

					hold := resp == "h"
					err = SimpleTimeout(func(ctx context.Context) error {
						return gc.ReplyToBid(ctx, hold)
					})
					if err != nil {
						l.Fatalw("failed to respond to bid",
							"err", err,
						)
					}
				}
			} else {
				if bs.AwaitingResponse {
					fmt.Printf("Waiting for player %d to respond\n", bs.Responder)
				} else {
					fmt.Printf("Waiting for player %d to make a call\n", bs.Caller)
				}
			}
		}
	case skat.PhaseDeclaration:
		{
			if gs.Declarer == st.PlayerIndex {
				declarationPhaseDeclarer(l, gc, st)
			} else {
				startView("Waiting for declarer...", gs.Hand)
			}
			endView()
		}
	case skat.PhasePlaying:
		{
			playingPhase(l, gc, st)
		}
	case skat.PhaseScored:
		{
			condition := "lost"
			if gs.LossReason == "" {
				condition = "won"
			}
			startView(fmt.Sprintf("The declarer has %s!", condition), nil)

			declarerPoints := gs.Players[gs.Declarer].WonCardPoints
			fmt.Printf(
				"Declarer points: %d\nDefender points: %d\n",
				declarerPoints,
				120-declarerPoints,
			)
			fmt.Printf(
				"Modifiers: %s\n",
				gs.FinalModifiers.Without(skat.AnnouncementModifiers).Pretty(),
			)
			fmt.Printf(
				"Announced: %s\n",
				gs.FinalModifiers.Without(skat.StateModifiers).Pretty(),
			)
			fmt.Printf(
				"Game value: %d\n",
				gs.FinalGameValue,
			)
			fmt.Printf(
				"Bid: %d\n",
				gs.LastBiddingCall,
			)
			fmt.Printf("\n")

			if gs.LossReason != "" {
				fmt.Printf("Declarer loss reason: ")
				switch gs.LossReason {
				case skat.LossReasonNotEnoughPoints:
					{
						fmt.Printf("Not enough points scored\n")
					}
				case skat.LossReasonOverbid:
					{
						fmt.Printf("Overbid\n")
					}
				case skat.LossReasonNoSchneider:
					{
						fmt.Printf("Schneider announced, but not achieved\n")
					}
				case skat.LossReasonNoSchwarz:
					{
						fmt.Printf("Schwarz announced, but not achieved\n")
					}
				case skat.LossReasonNotNull:
					{
						fmt.Printf("Not a null game\n")
					}
				default:
					{
						fmt.Printf("?!\n")
					}
				}
			}

			if (gs.Declarer != st.PlayerIndex && gs.LossReason != "") || (gs.Declarer == st.PlayerIndex && gs.LossReason == "") {
				fmt.Printf("Congratulations!\n")
			} else {
				fmt.Printf("Better luck next time!\n")
			}

			endView()
		}
	}
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}
	zap.ReplaceGlobals(logger)

	flag.Parse()

	clientID := flag.Arg(0)
	clientSecret := flag.Arg(1)

	sl := zap.S()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := singleuser.DialAddr(
		sl.With("component", "net_client"),
		ctx,
		*serverAddress,
		&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         "skat-server",
			NextProtos: []string{
				singleuser.ProtocolName,
			},
		},
	)
	if err != nil {
		sl.Fatalw("failed to connect to server",
			"err", err,
		)
	}
	defer conn.Close()

	gc, err := singleuser.NewGameClient(
		sl.With("component", "game_client"),
		ctx,
		conn,
	)
	if err != nil {
		sl.Fatalw("failed to connect to server",
			"err", err,
		)
	}
	defer gc.Close()

	sl.Infow("connected")

	err = gc.Login(ctx, clientID, clientSecret, *serverPassword)
	if err != nil {
		sl.Fatalw(
			"failed to login",
			"err", err,
		)
	}
	sl.Infow("login successful")
	cancel()

	err = SimpleTimeout(func(ctx context.Context) error {
		return gc.NetPing(ctx)
	})
	if err != nil {
		sl.Fatalw("failed to ping",
			"err", err,
		)
	}

	for {
		state, ok := <-gc.StateChannel()
		if !ok {
			break
		}
		HandleGameState(sl, gc, state)
	}
}
