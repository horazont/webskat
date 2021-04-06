package skat

import (
	"errors"
	"strings"
)

type GameType int

var (
	ErrCardNotPresent     = errors.New("the card is not present")
	ErrCardAlreadyPresent = errors.New("the card is already in the set")
)

const (
	GameTypeDiamonds GameType = 1
	GameTypeHearts   GameType = 2
	GameTypeSpades   GameType = 3
	GameTypeClubs    GameType = 4
	GameTypeGrand    GameType = 5
	GameTypeNull     GameType = 6
	GameTypeJunk     GameType = 7
)

var (
	StandardGameTypes = []GameType{GameTypeDiamonds, GameTypeHearts, GameTypeSpades, GameTypeClubs, GameTypeGrand, GameTypeNull}
	SuitGameTypes     = []GameType{GameTypeDiamonds, GameTypeHearts, GameTypeSpades, GameTypeClubs}
)

type GameModifier uint16

const (
	NoGameModifiers                GameModifier = 0
	GameModifierHand               GameModifier = 1 << 0
	GameModifierSchneider          GameModifier = 1 << 1
	GameModifierSchwarz            GameModifier = 1 << 2
	GameModifierSchneiderAnnounced GameModifier = 1 << 3
	GameModifierSchwarzAnnounced   GameModifier = 1 << 4
	GameModifierOuvert             GameModifier = 1 << 5

	StateModifiers        GameModifier = GameModifierHand | GameModifierSchneider | GameModifierSchwarz
	AnnouncementModifiers GameModifier = GameModifierSchneiderAnnounced | GameModifierSchwarzAnnounced | GameModifierOuvert
)

type CardType int

const (
	Card7     CardType = 0
	Card8     CardType = 1
	Card9     CardType = 2
	CardQueen CardType = 5
	CardKing  CardType = 6
	Card10    CardType = 7
	CardAce   CardType = 8
	CardJack  CardType = 9

	// must be higher than any RelativePower() value returned; as most of
	// RelativePower consists of int(CardType), the constant is defined here
	trumpRelativePowerOffset = 100
)

var (
	CardTypes = [8]CardType{Card7, Card8, Card9, CardQueen, CardKing, Card10, CardAce, CardJack}
)

type Suit int

const (
	SuitDiamonds Suit = 0
	SuitHearts   Suit = 1
	SuitSpades   Suit = 2
	SuitClubs    Suit = 3
)

var (
	Suits = [4]Suit{SuitDiamonds, SuitHearts, SuitSpades, SuitClubs}
)

type EffectiveSuit int

const (
	EffectiveSuitDiamonds EffectiveSuit = 0
	EffectiveSuitHearts   EffectiveSuit = 1
	EffectiveSuitSpades   EffectiveSuit = 2
	EffectiveSuitClubs    EffectiveSuit = 3
	EffectiveSuitTrumps   EffectiveSuit = 4
)

var (
	baseEffectiveSuitMap = map[Suit]EffectiveSuit{
		SuitDiamonds: EffectiveSuitDiamonds,
		SuitHearts:   EffectiveSuitHearts,
		SuitSpades:   EffectiveSuitSpades,
		SuitClubs:    EffectiveSuitClubs,
	}
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

func (c CardType) Pretty() string {
	switch c {
	case Card7:
		return "7"
	case Card8:
		return "8"
	case Card9:
		return "9"
	case Card10:
		return "0"
	case CardJack:
		return "J"
	case CardQueen:
		return "Q"
	case CardKing:
		return "K"
	case CardAce:
		return "A"
	}
	return "-"
}

func (c CardType) As(suit Suit) Card {
	return Card{c, suit}
}

func (s Suit) As(c CardType) Card {
	return Card{c, s}
}

func (s Suit) Pretty() string {
	switch s {
	case SuitDiamonds:
		return "♦"
	case SuitHearts:
		return "♥"
	case SuitSpades:
		return "♠"
	case SuitClubs:
		return "♣"
	}
	return "-"
}

type Card struct {
	Type CardType
	Suit Suit
}

func (c Card) Value() int {
	return c.Type.Value()
}

func (c Card) Pretty() string {
	return c.Suit.Pretty() + c.Type.Pretty()
}

func (c Card) EffectiveSuit(gameType GameType) EffectiveSuit {
	switch gameType {
	case GameTypeGrand:
		if c.Type == CardJack {
			return EffectiveSuitTrumps
		}
	case GameTypeDiamonds:
		if c.Suit == SuitDiamonds || c.Type == CardJack {
			return EffectiveSuitTrumps
		}
	case GameTypeHearts:
		if c.Suit == SuitHearts || c.Type == CardJack {
			return EffectiveSuitTrumps
		}
	case GameTypeSpades:
		if c.Suit == SuitSpades || c.Type == CardJack {
			return EffectiveSuitTrumps
		}
	case GameTypeClubs:
		if c.Suit == SuitClubs || c.Type == CardJack {
			return EffectiveSuitTrumps
		}
	}
	return baseEffectiveSuitMap[c.Suit]
}

func (c Card) RelativePower(gameType GameType) int {
	switch gameType {
	case GameTypeGrand:
		if c.Type == CardJack {
			return int(c.Suit)
		}
		return int(c.Type)
	case GameTypeNull:
		switch c.Type {
		case Card10:
			return 3
		case CardJack:
			return 4
		}
		return int(c.Type)
	case GameTypeDiamonds, GameTypeHearts, GameTypeSpades, GameTypeClubs:
		suit := c.EffectiveSuit(gameType)
		if suit == EffectiveSuitTrumps && c.Type == CardJack {
			return int(c.Suit) + trumpRelativePowerOffset
		}
		return int(c.Type)
	}
	return 0
}

type Trick [3]Card

func (t Trick) Value() int {
	return t[0].Value() + t[1].Value() + t[2].Value()
}

func (t Trick) Copy() (result Trick) {
	copy(result[:], t[:])
	return result
}

func (t Trick) AsCardSet() CardSet {
	return CardSet{t[0], t[1], t[2]}
}

func (t Trick) EffectiveSuit(gameType GameType) EffectiveSuit {
	return t[0].EffectiveSuit(gameType)
}

func (t Trick) Taker(gameType GameType) (winner int) {
	suit := t.EffectiveSuit(gameType)
	highestPower := -1
	winner = -1
	for i, card := range t {
		cardSuit := card.EffectiveSuit(gameType)
		cardPower := 0
		if cardSuit == EffectiveSuitTrumps {
			cardPower = cardPower + trumpRelativePowerOffset
		} else if cardSuit != suit {
			continue
		}
		cardPower = cardPower + card.RelativePower(gameType)
		if cardPower > highestPower {
			highestPower = cardPower
			winner = i
		}
	}
	return winner
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

func (m GameModifier) Test(other GameModifier) bool {
	return m&other == other
}

func (m GameModifier) With(other GameModifier) GameModifier {
	return m | other
}

func (m GameModifier) Without(other GameModifier) GameModifier {
	return m &^ other
}

func (m GameModifier) IsAnnounceable() bool {
	return m.Normalized() == m && m&^AnnouncementModifiers == 0
}

// Include implicit modifiers
func (modifiers GameModifier) Normalized() GameModifier {
	result := modifiers
	if result.Test(GameModifierSchwarz) {
		result = result.With(GameModifierSchneider)
	}
	if result.Test(GameModifierSchwarzAnnounced) {
		result = result.With(GameModifierSchneiderAnnounced)
	}
	return result
}

// Test whether the given modifier set is valid for an announcement
func (modifiers GameModifier) ValidForGame(game GameType) bool {
	if modifiers != modifiers.Normalized() {
		return false
	}
	switch game {
	case GameTypeNull:
		{
			if modifiers.Test(GameModifierSchneiderAnnounced) || modifiers.Test(GameModifierSchwarzAnnounced) {
				return false
			}
			return true
		}
	// Suit games + Grand
	case GameTypeClubs, GameTypeDiamonds, GameTypeHearts, GameTypeSpades, GameTypeGrand:
		{
			if modifiers.Test(GameModifierSchneiderAnnounced) && !modifiers.Test(GameModifierHand) {
				return false
			}
			return true
		}
	}
	return false
}

func (cs CardSet) Contains(c Card) bool {
	for _, cin := range cs {
		if cin == c {
			return true
		}
	}
	return false
}

func (cs CardSet) Copy() CardSet {
	result := make([]Card, len(cs))
	copy(result, cs)
	return result
}

func (cs CardSet) Pop(c Card) (CardSet, error) {
	result := make([]Card, len(cs)-1)
	for i, cin := range cs {
		if cin == c {
			copy(result[:i], cs[:i])
			copy(result[i:], cs[i+1:])
			return result, nil
		}
	}
	return cs, ErrCardNotPresent
}

func (cs CardSet) Push(c Card) (CardSet, error) {
	result := make([]Card, len(cs)+1)
	for _, cin := range cs {
		if cin == c {
			return nil, ErrCardAlreadyPresent
		}
	}
	copy(result, cs)
	result[len(cs)] = c
	return result, nil
}

func (cs CardSet) Value() (sum int) {
	for _, card := range cs {
		sum = sum + card.Value()
	}
	return sum
}

func (cs CardSet) Pretty() string {
	var sb strings.Builder
	for i, card := range cs {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(card.Pretty())
	}
	return sb.String()
}
