package skat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameModifierEnumIsPowerOfTwoBased(t *testing.T) {
	assert.Equal(t, GameModifier(1), GameModifierHand)
	assert.Equal(t, GameModifier(2), GameModifierSchneider)
	assert.Equal(t, GameModifier(4), GameModifierSchwarz)
	assert.Equal(t, GameModifier(8), GameModifierOuvert)
}

func TestGameTypeEnum(t *testing.T) {
	assert.Equal(t, GameType(1), GameTypeBells)
	assert.Equal(t, GameType(2), GameTypeHearts)
	assert.Equal(t, GameType(3), GameTypeLeaves)
	assert.Equal(t, GameType(4), GameTypeAcorns)
}

func TestCardTypeValue(t *testing.T) {
	assert.Equal(t, Card7.Value(), 0)
	assert.Equal(t, Card8.Value(), 0)
	assert.Equal(t, Card9.Value(), 0)
	assert.Equal(t, CardQueen.Value(), 3)
	assert.Equal(t, CardKing.Value(), 4)
	assert.Equal(t, Card10.Value(), 10)
	assert.Equal(t, CardAce.Value(), 11)
	assert.Equal(t, CardJack.Value(), 2)
}

func TestGameModifierNormalized(t *testing.T) {
	t.Run("no modifiers pass unmodified", func(t *testing.T) {
		assert.Equal(t, NoGameModifiers, NoGameModifiers.Normalized())
	})

	t.Run("include Schneider in Schwarz announcement", func(t *testing.T) {
		assert.Equal(t, GameModifierSchneider|GameModifierSchwarz, GameModifierSchwarz.Normalized())
	})

	t.Run("Schneider passes unmodified", func(t *testing.T) {
		assert.Equal(t, GameModifierSchneider, GameModifierSchneider.Normalized())
	})

	t.Run("Hand passes unmodified", func(t *testing.T) {
		assert.Equal(t, GameModifierHand, GameModifierHand.Normalized())
	})

	t.Run("Ouvert passes unmodified", func(t *testing.T) {
		assert.Equal(t, GameModifierOuvert, GameModifierOuvert.Normalized())
	})
}

func TestGameModifierCanBeAnnouncedForGame(t *testing.T) {
	suitGames := []GameType{GameTypeAcorns, GameTypeBells, GameTypeHearts, GameTypeLeaves, GameTypeGrand}

	t.Run("announcement without modifiers is allowed for all game types", func(t *testing.T) {
		for _, game := range StandardGameTypes {
			assert.True(t, NoGameModifiers.CanBeAnnouncedForGame(game, NoGameModifiers))
		}
	})

	t.Run("no modifiers are allowed for Junk", func(t *testing.T) {
		modifiers := []GameModifier{
			GameModifierHand,
			GameModifierSchneider,
			GameModifierSchwarz,
			GameModifierOuvert,
		}
		for _, modifier := range modifiers {
			assert.False(t, modifier.CanBeAnnouncedForGame(GameTypeJunk, NoGameModifiers))
		}
	})

	t.Run("Hand can not be announced for anything", func(t *testing.T) {
		for _, game := range StandardGameTypes {
			assert.False(t, GameModifierHand.CanBeAnnouncedForGame(game, NoGameModifiers))
		}
	})

	t.Run("Schneider can be announced for suit games with Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.True(t, GameModifierSchneider.CanBeAnnouncedForGame(game, GameModifierHand))
		}
	})

	t.Run("Schwarz+Schneider can be announced for suit games with Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.True(t, (GameModifierSchneider|GameModifierSchwarz).CanBeAnnouncedForGame(game, GameModifierHand))
		}
	})

	t.Run("Ouvert+Schwarz+Schneider can be announced for suit games with Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.True(t, (GameModifierSchneider|GameModifierSchwarz|GameModifierOuvert).CanBeAnnouncedForGame(game, GameModifierHand))
		}
	})

	t.Run("Schneider cannot be announced for suit games without Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.False(t, GameModifierSchneider.CanBeAnnouncedForGame(game, NoGameModifiers))
		}
	})

	t.Run("Schwarz+Schneider cannot be announced for suit games without Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.False(t, (GameModifierSchneider|GameModifierSchwarz).CanBeAnnouncedForGame(game, NoGameModifiers))
		}
	})

	t.Run("Ouvert+Schwarz+Schneider cannot be announced for suit games without Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.False(t, (GameModifierSchneider|GameModifierSchwarz|GameModifierOuvert).CanBeAnnouncedForGame(game, NoGameModifiers))
		}
	})

	t.Run("Schneider can not be announced for null game", func(t *testing.T) {
		assert.False(t, GameModifierSchneider.CanBeAnnouncedForGame(GameTypeNull, NoGameModifiers))
	})

	t.Run("Schwarz can not be announced for null game", func(t *testing.T) {
		assert.False(t, GameModifierSchwarz.CanBeAnnouncedForGame(GameTypeNull, NoGameModifiers))
	})

	t.Run("Ouvert can be announced for null game with Hand", func(t *testing.T) {
		assert.True(t, GameModifierOuvert.CanBeAnnouncedForGame(GameTypeNull, GameModifierHand))
	})

	t.Run("Ouvert can be announced for null game without Hand", func(t *testing.T) {
		assert.True(t, GameModifierOuvert.CanBeAnnouncedForGame(GameTypeNull, NoGameModifiers))
	})
}

func TestCardOperations(t *testing.T) {
	t.Run("contains", func(t *testing.T) {
		cards := make(CardSet, 0)
		testCard := Card{Card10, SuitHearts}
		assert.False(t, cards.Contains(testCard))
	})

	t.Run("pop", func(t *testing.T) {
		cards := NewCardDeck()
		toPop := Card{Card10, SuitHearts}
		cardsAfter, err := cards.Pop(toPop)
		assert.Nil(t, err)
		assert.False(t, cardsAfter.Contains(toPop))
		assert.True(t, cardsAfter.Contains(Card{Card9, SuitHearts}))
		assert.True(t, cardsAfter.Contains(Card{Card10, SuitBells}))
	})

	t.Run("pop returns no such card if not in set", func(t *testing.T) {
		cards := CardSet{
			Card{Card9, SuitHearts},
		}
		cardsBefore := make(CardSet, len(cards))
		copy(cardsBefore, cards)

		toPop := Card{Card10, SuitHearts}
		cardsAfter, err := cards.Pop(toPop)
		assert.Equal(t, ErrCardNotPresent, err)
		assert.Equal(t, cardsBefore, cardsAfter)
	})
}
