package skat

import (
	"errors"
)

var (
	ErrMustFollowSuit = errors.New("must follow suit")
)

type PlayingPlayerState struct {
	Hand     CardSet
	WonCards CardSet
}

type PlayingState struct {
	forehand        int
	current         int
	declarer        int
	gameType        GameType
	lastTrick       Trick
	lastTrickWinner int
	table           CardSet
	players         [3]PlayingPlayerState
}

func NewPlayingState(declarer int, gameType GameType, hands [3]*CardSet, pushedCards CardSet) *PlayingState {
	result := &PlayingState{
		forehand:        PlayerInitialForehand,
		current:         PlayerInitialForehand,
		declarer:        declarer,
		gameType:        gameType,
		lastTrickWinner: PlayerNone,
		players: [3]PlayingPlayerState{
			PlayingPlayerState{
				Hand:     hands[0].Copy(),
				WonCards: make(CardSet, 0),
			},
			PlayingPlayerState{
				Hand:     hands[1].Copy(),
				WonCards: make(CardSet, 0),
			},
			PlayingPlayerState{
				Hand:     hands[2].Copy(),
				WonCards: make(CardSet, 0),
			},
		},
	}
	if len(pushedCards) > 0 {
		result.players[declarer].WonCards = pushedCards.Copy()
	}
	return result
}

func (s *PlayingState) Declarer() int {
	return s.declarer
}

func (s *PlayingState) GameType() GameType {
	return s.gameType
}

func (s *PlayingState) GetLastTrick() (Trick, int) {
	return s.lastTrick.Copy(), s.lastTrickWinner
}

func (s *PlayingState) GetWonCards(player int) CardSet {
	return s.players[player].WonCards.Copy()
}

func (s *PlayingState) GetCurrentPlayer() int {
	return s.current
}

func (s *PlayingState) tableSuit() EffectiveSuit {
	return s.table[0].EffectiveSuit(s.gameType)
}

func (s *PlayingState) Play(player int, card Card) (err error) {
	if player != s.current {
		return ErrNotYourTurn
	}

	hand := s.players[player].Hand

	if len(s.table) > 0 {
		tableSuit := s.tableSuit()
		cardSuit := card.EffectiveSuit(s.gameType)
		// If the effective suit of the card does not match what’s on the
		// table...
		if cardSuit != tableSuit {
			for _, card := range hand {
				if card.EffectiveSuit(s.gameType) == tableSuit {
					// ... we cannot allow it if there is any possible card
					// to be played.
					return ErrMustFollowSuit
				}
			}
		}
	}

	newHand, err := hand.Pop(card)
	if err != nil {
		return err
	}
	newTable, err := s.table.Push(card)
	if err != nil {
		return err
	}
	s.players[player].Hand = newHand
	s.table = newTable

	if s.current == PlayerInitialRearhand {
		s.current = PlayerInitialForehand
	} else {
		s.current = s.current + 1
	}
	if len(s.table) == 3 {
		s.concludeTrick()
	}
	return err
}

func (s *PlayingState) relativeToAbsolutePlayer(relativePlayer int) int {
	return (relativePlayer + s.forehand) % 3
}

func (s *PlayingState) concludeTrick() {
	s.lastTrick = Trick{s.table[0], s.table[1], s.table[2]}
	s.lastTrickWinner = s.relativeToAbsolutePlayer(s.lastTrick.Taker(s.gameType))
	s.players[s.lastTrickWinner].WonCards = append(s.players[s.lastTrickWinner].WonCards, s.table...)
	s.table = s.table[:0]
	s.current = s.lastTrickWinner
	s.forehand = s.lastTrickWinner
}

func (s *PlayingState) GetTable() CardSet {
	return s.table.Copy()
}

func (s *PlayingState) GetHand(player int) CardSet {
	return s.players[player].Hand.Copy()
}
