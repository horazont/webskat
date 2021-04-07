package skat

import (
	"crypto/rand"
	"errors"
)

var (
	ErrMissingSeed     = errors.New("not all players have submitted a seed")
	ErrWrongPhase      = errors.New("wrong game phase for this action")
	ErrNotYourTurn     = errors.New("this is not your turn")
	ErrBiddingNotDone  = errors.New("bidding has not completed yet")
	ErrInvalidGameType = errors.New("invalid game type")
	ErrNotImplemented  = errors.New("not implemented")
	ErrInvalidGame     = errors.New("invalid game")
	ErrInvalidPush     = errors.New("invalid push request")
)

const (
	PlayerNone              = -1
	PlayerInitialForehand   = 0
	PlayerInitialMiddlehand = 1
	PlayerInitialRearhand   = 2
)

type GamePhase int

const (
	// Collection of seeds
	PhaseInit GamePhase = 0

	// Cards have been dealt, players bid
	PhaseBidding GamePhase = 1

	// Bidding is over, waiting for declaring player to make a declaration
	PhaseDeclaration GamePhase = 2

	// Play time
	PhasePlaying GamePhase = 3

	// All cards have been played in one way or another, game has been scored
	PhaseScored GamePhase = 4
)

const (
	ServerSeedSize = 16
)

type CommonPlayerState struct {
	Seed  []byte
	Hand  CardSet
	Score int
}

type GameState struct {
	phase               GamePhase
	withDealer          bool
	serverSeed          []byte
	dealerSeed          []byte
	dealerLookingAtHand int
	scoring             ScoreDefinition

	skat       CardSet
	players    [3]CommonPlayerState
	modifiers  GameModifier
	lossReason string

	biddingState *BiddingState
	playingState *PlayingState
}

func NewGame(withDealer bool, scoring *ScoreDefinition) *GameState {
	return &GameState{
		withDealer:          withDealer,
		phase:               PhaseInit,
		dealerLookingAtHand: PlayerNone,
		scoring:             *scoring,
		modifiers:           GameModifierHand,
	}
}

func (g *GameState) Phase() GamePhase {
	return g.phase
}

func (g *GameState) SetSeed(player int, seed []byte) error {
	if g.phase != PhaseInit {
		return ErrWrongPhase
	}
	g.players[player].Seed = seed
	return nil
}

func (g *GameState) SetDealerSeed(seed []byte) error {
	if g.phase != PhaseInit {
		return ErrWrongPhase
	}
	if g.withDealer {
		g.dealerSeed = seed
	}
	return nil
}

func (g *GameState) GenerateServerSeed() error {
	seed := make([]byte, ServerSeedSize)
	_, err := rand.Read(seed)
	if err != nil {
		return err
	}
	return g.SetServerSeed(seed)
}

func (g *GameState) SetServerSeed(seed []byte) error {
	if g.phase != PhaseInit {
		return ErrWrongPhase
	}
	if g.serverSeed != nil {
		return ErrWrongPhase
	}
	g.serverSeed = seed

	return nil
}

func (g *GameState) ComposedSeed() ([]byte, error) {
	if g.serverSeed == nil {
		return nil, ErrMissingSeed
	}
	result := make([]byte, len(g.serverSeed))
	copy(result, g.serverSeed)
	for _, player := range g.players {
		if player.Seed == nil {
			return nil, ErrMissingSeed
		}
		result = append(result, player.Seed...)
	}
	if g.withDealer && g.dealerSeed == nil {
		return nil, ErrMissingSeed
	}
	result = append(result, g.dealerSeed...)
	return result, nil
}

func (g *GameState) dealRoundOfHands(cards []Card, n int) ([]Card, error) {
	var err error
	for i := 0; i < 3; i = i + 1 {
		var dealt []Card
		cards, dealt, err = DrawCards(cards, n)
		if err != nil {
			return nil, err
		}
		g.players[i].Hand = append(g.players[i].Hand, dealt...)
	}

	return cards, nil
}

// Transition PhaseInit -> PhaseBidding
func (g *GameState) Deal() error {
	if g.phase != PhaseInit {
		return ErrWrongPhase
	}

	seed, err := g.ComposedSeed()
	if err != nil {
		return err
	}

	deck := NewCardDeck()
	err = ShuffleDeckWithSeed(seed, &deck)
	if err != nil {
		return err
	}

	deck, err = g.dealRoundOfHands(deck, 3)
	if err != nil {
		return err
	}
	deck, g.skat, err = DrawCards(deck, 2)
	if err != nil {
		return err
	}
	deck, err = g.dealRoundOfHands(deck, 4)
	if err != nil {
		return err
	}
	deck, err = g.dealRoundOfHands(deck, 3)
	if err != nil {
		return err
	}
	if len(deck) != 0 {
		panic("too many cards in generated deck")
	}

	g.initBidding()
	return nil
}

func (g *GameState) initBidding() {
	g.phase = PhaseBidding
	g.biddingState = NewBiddingState()
}

func (g *GameState) Bidding() *BiddingState {
	if g.phase != PhaseBidding {
		return nil
	}
	return g.biddingState
}

// Transition PhaseBidding -> PhaseDeclaration
func (g *GameState) ConcludeBidding() error {
	if g.phase != PhaseBidding {
		return ErrWrongPhase
	}
	if !g.biddingState.Done() {
		return ErrBiddingNotDone
	}

	// TODO: if nobody won the bidding, either transition to Junk or abort
	// game
	g.initDeclaring()
	return nil
}

func (g *GameState) initDeclaring() {
	g.phase = PhaseDeclaration
}

func (g *GameState) TakeSkat(player int) error {
	if g.phase != PhaseDeclaration {
		return ErrWrongPhase
	}
	declarer := g.biddingState.Declarer()
	if declarer != player {
		return ErrNotYourTurn
	}
	if !g.modifiers.Test(GameModifierHand) {
		return ErrWrongPhase
	}
	g.modifiers = g.modifiers.Without(GameModifierHand)
	g.players[declarer].Hand = append(g.players[declarer].Hand, g.skat...)
	return nil
}

func (g *GameState) Declare(player int, gameType GameType, announcedModifiers GameModifier, cardsToPush CardSet) error {
	if g.phase != PhaseDeclaration {
		return ErrWrongPhase
	}
	declarer := g.biddingState.Declarer()
	if declarer != player {
		return ErrNotYourTurn
	}
	if !announcedModifiers.IsAnnounceable() {
		return ErrInvalidGame
	}
	newModifiers := g.modifiers | announcedModifiers
	if !newModifiers.ValidForGame(gameType) {
		return ErrInvalidGame
	}

	if !g.modifiers.Test(GameModifierHand) && len(cardsToPush) != 2 {
		return ErrInvalidPush
	}
	if g.modifiers.Test(GameModifierHand) && len(cardsToPush) != 0 {
		return ErrInvalidPush
	}
	newHand := g.players[player].Hand
	for _, card := range cardsToPush {
		var err error
		newHand, err = newHand.Pop(card)
		if err != nil {
			return ErrInvalidPush
		}
	}

	var skatCards CardSet
	if len(cardsToPush) > 0 {
		skatCards = cardsToPush
	} else {
		skatCards = g.skat
	}

	g.players[player].Hand = newHand
	g.phase = PhasePlaying
	g.playingState = NewPlayingState(
		g.biddingState.Declarer(),
		gameType,
		[3]*CardSet{
			&g.players[0].Hand,
			&g.players[1].Hand,
			&g.players[2].Hand,
		},
		skatCards,
	)
	// now for the declarer, we add the skat to the hand for post-game
	// evaluation
	g.players[player].Hand, _ = g.players[player].Hand.Push(
		g.skat[0],
	)
	g.players[player].Hand, _ = g.players[player].Hand.Push(
		g.skat[1],
	)
	return nil
}

func (g *GameState) GetHand(player int) CardSet {
	if g.phase == PhasePlaying {
		return g.playingState.GetHand(player)
	}
	return g.players[player].Hand.Copy()
}

func (g *GameState) GetSkat() CardSet {
	return g.skat.Copy()
}

func (g *GameState) Modifiers() GameModifier {
	return g.modifiers
}

func (g *GameState) Playing() *PlayingState {
	return g.playingState
}

func (g *GameState) GetScore(player int) int {
	return g.players[player].Score
}

func (g *GameState) GetLossReason() string {
	return g.lossReason
}

func (g *GameState) EvaluateGame() error {
	if g.phase != PhasePlaying {
		return ErrWrongPhase
	}
	if len(g.playingState.GetHand(PlayerInitialForehand)) > 0 {
		return ErrWrongPhase
	}
	declarer := g.biddingState.Declarer()
	resultModifiers, declarerScore, _ := EvaluateWonCards(
		[3]CardSet{
			g.playingState.GetWonCards(PlayerInitialForehand),
			g.playingState.GetWonCards(PlayerInitialMiddlehand),
			g.playingState.GetWonCards(PlayerInitialRearhand),
		},
		declarer,
	)
	modifiers := g.modifiers | resultModifiers
	baseValue, factor := CalculateGameValue(
		g.players[declarer].Hand,
		g.playingState.GameType(),
		modifiers,
	)
	declarerWon, gameValue, lossReason := EvaluateGame(
		baseValue,
		factor,
		declarerScore,
		g.biddingState.CalledGameValue(),
		g.playingState.GameType(),
		modifiers,
	)
	playerScores := g.scoring.CalculateScore(
		gameValue,
		declarer,
		declarerWon,
	)
	g.modifiers = modifiers
	for i := range g.players {
		g.players[i].Score = playerScores[i]
	}
	g.lossReason = lossReason
	g.phase = PhaseScored
	return nil
}
