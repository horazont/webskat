package skat

import (
	"errors"
)

type GameType int

var (
	ErrCardNotPresent = errors.New("the card is not present")
)

const (
	GameTypeBells  GameType = 1
	GameTypeHearts GameType = 2
	GameTypeLeaves GameType = 3
	GameTypeAcorns GameType = 4
	GameTypeGrand  GameType = 5
	GameTypeNull   GameType = 6
	GameTypeJunk   GameType = 7
)

var (
	StandardGameTypes = []GameType{GameTypeBells, GameTypeHearts, GameTypeLeaves, GameTypeAcorns, GameTypeGrand, GameTypeNull}
)

type GameModifier uint16

const (
	NoGameModifiers       GameModifier = 0
	GameModifierHand      GameModifier = 1 << 0
	GameModifierSchneider GameModifier = 1 << 1
	GameModifierSchwarz   GameModifier = 1 << 2
	GameModifierOuvert    GameModifier = 1 << 3
)

type CardType byte

const (
	Card7     CardType = '7'
	Card8     CardType = '8'
	Card9     CardType = '9'
	CardQueen CardType = 'Q'
	CardKing  CardType = 'K'
	Card10    CardType = '0'
	CardAce   CardType = 'A'
	CardJack  CardType = 'J'
)

var (
	CardTypes = [8]CardType{Card7, Card8, Card9, CardQueen, CardKing, Card10, CardAce, CardJack}
)

type Suit byte

const (
	SuitBells  Suit = 'b'
	SuitHearts Suit = 'h'
	SuitLeaves Suit = 'l'
	SuitAcorns Suit = 'a'
)

var (
	Suits = [4]Suit{SuitBells, SuitHearts, SuitLeaves, SuitAcorns}
)

func (c CardType) Value() int {
	switch c {
	case CardQueen:
		return 3
	case CardKing:
		return 4
	case Card10:
		return 10
	case CardAce:
		return 11
	case CardJack:
		return 2
	}
	return 0
}

type Card struct {
	Type CardType
	Suit Suit
}

func (c Card) Value() int {
	return c.Value()
}

type Trick [3]Card

func (t Trick) Value() int {
	return t[0].Value() + t[1].Value() + t[2].Value()
}

type CardSet []Card

func NewCardDeck() (result CardSet) {
	result = make([]Card, 32)
	iout := 0
	for _, suit := range Suits {
		for _, type_ := range CardTypes {
			result[iout] = Card{type_, suit}
			iout = iout + 1
		}
	}
	return result
}

// Test whether the given modifier set is valid for an announcement
func (modifiers GameModifier) CanBeAnnouncedForGame(
	game GameType,
	// Modifiers implicit from the current game state
	stateModifiers GameModifier,
) bool {
	if modifiers&GameModifierHand == GameModifierHand {
		// Hand is not announced, it is an ambient
		return false
	}
	switch game {
	case GameTypeNull:
		{
			if modifiers&GameModifierSchneider == GameModifierSchneider {
				return false
			}
			if modifiers&GameModifierSchwarz == GameModifierSchwarz {
				return false
			}
			return true
		}
	// Suit games + Grand
	case GameTypeAcorns, GameTypeBells, GameTypeHearts, GameTypeLeaves, GameTypeGrand:
		{
			if stateModifiers&GameModifierHand == 0 && modifiers != 0 {
				return false
			}
			return true
		}
	}
	return false
}

// Include implicit modifiers
func (modifiers GameModifier) Normalized() GameModifier {
	result := modifiers
	if result&GameModifierSchwarz == GameModifierSchwarz {
		result = result | GameModifierSchneider
	}
	return result
}

func (cs CardSet) Contains(c Card) bool {
	for _, cin := range cs {
		if cin == c {
			return true
		}
	}
	return false
}

func (cs CardSet) Pop(c Card) (CardSet, error) {
	for i, cin := range cs {
		if cin == c {
			return append(cs[:i], cs[i+1:]...), nil
		}
	}
	return cs, ErrCardNotPresent
}
