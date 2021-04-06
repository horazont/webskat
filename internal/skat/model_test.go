package skat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGameModifier(t *testing.T) {
	t.Run("EnumIsPowerOfTwoBased", func(t *testing.T) {
		assert.Equal(t, GameModifier(1), GameModifierHand)
		assert.Equal(t, GameModifier(2), GameModifierSchneider)
		assert.Equal(t, GameModifier(4), GameModifierSchwarz)
		assert.Equal(t, GameModifier(8), GameModifierSchneiderAnnounced)
		assert.Equal(t, GameModifier(16), GameModifierSchwarzAnnounced)
		assert.Equal(t, GameModifier(32), GameModifierOuvert)
	})

	t.Run("AnnouncedGameModifiers masks announcements", func(t *testing.T) {
		assert.Equal(t, GameModifierSchneiderAnnounced|GameModifierSchwarzAnnounced|GameModifierOuvert, AnnouncementModifiers)
	})

	t.Run("StateModifiers masks announcements", func(t *testing.T) {
		assert.Equal(t, GameModifierHand|GameModifierSchneider|GameModifierSchwarz, StateModifiers)
	})

	t.Run("bitops: has", func(t *testing.T) {
		assert.True(t, GameModifierHand.Test(GameModifierHand))
		assert.False(t, GameModifierHand.Test(GameModifierSchneider))
		assert.False(t, GameModifierSchneider.Test(GameModifierHand))
	})

	t.Run("bitops: with", func(t *testing.T) {
		assert.Equal(t, GameModifierHand, GameModifierHand.With(GameModifierHand))
		assert.Equal(t, GameModifierHand, NoGameModifiers.With(GameModifierHand))
	})

	t.Run("bitops: without", func(t *testing.T) {
		assert.Equal(t, NoGameModifiers, GameModifierHand.Without(GameModifierHand))
		assert.Equal(t, NoGameModifiers, NoGameModifiers.Without(GameModifierHand))
	})

	t.Run("normalized", func(t *testing.T) {
		// Schwarz implies Schneider
		assert.Equal(t, (GameModifierSchneider | GameModifierSchwarz), GameModifierSchwarz.Normalized())
		// SchwarzAnnounced implies SchneiderAnnounced
		assert.Equal(t, (GameModifierSchneiderAnnounced | GameModifierSchwarzAnnounced), GameModifierSchwarzAnnounced.Normalized())

		// nothing else implies SchneiderAnnounced or Schneider
		for _, mod := range []GameModifier{NoGameModifiers, GameModifierHand, GameModifierSchneider, GameModifierSchneiderAnnounced, GameModifierOuvert} {
			assert.Equal(t, mod, mod.Normalized())
		}
	})

	t.Run("IsAnnounceable", func(t *testing.T) {
		assert.True(t, NoGameModifiers.IsAnnounceable())
		assert.True(t, GameModifierSchneiderAnnounced.IsAnnounceable())
		assert.True(t, (GameModifierSchneiderAnnounced | GameModifierSchwarzAnnounced).IsAnnounceable())
		assert.True(t, GameModifierOuvert.IsAnnounceable())

		assert.False(t, GameModifierSchwarzAnnounced.IsAnnounceable())
		assert.False(t, GameModifierHand.IsAnnounceable())
		assert.False(t, (GameModifierHand | GameModifierSchneiderAnnounced).IsAnnounceable())
	})
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

func TestGameModifierValidForGame(t *testing.T) {
	suitGames := []GameType{GameTypeAcorns, GameTypeBells, GameTypeHearts, GameTypeLeaves, GameTypeGrand}

	t.Run("announcement without modifiers is allowed for all game types", func(t *testing.T) {
		for _, game := range StandardGameTypes {
			assert.True(t, NoGameModifiers.ValidForGame(game))
		}
	})

	t.Run("no modifiers are allowed for Junk", func(t *testing.T) {
		modifiers := []GameModifier{
			GameModifierHand,
			GameModifierSchneider,
			GameModifierSchneiderAnnounced,
			GameModifierSchwarz,
			GameModifierSchwarzAnnounced,
			GameModifierOuvert,
		}
		for _, modifier := range modifiers {
			assert.False(t, modifier.ValidForGame(GameTypeJunk))
		}
	})

	t.Run("Schneider can be announced for suit games with Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.True(t, (GameModifierSchneiderAnnounced | GameModifierHand).ValidForGame(game))
		}
	})

	t.Run("Schwarz+Schneider can be announced for suit games with Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.True(t, (GameModifierHand | GameModifierSchneiderAnnounced | GameModifierSchwarzAnnounced).ValidForGame(game))
		}
	})

	t.Run("Ouvert+Schwarz+Schneider can be announced for suit games with Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.True(t, (GameModifierHand | GameModifierSchneiderAnnounced | GameModifierSchwarzAnnounced | GameModifierOuvert).ValidForGame(game))
		}
	})

	t.Run("Schneider cannot be announced for suit games without Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.False(t, GameModifierSchneiderAnnounced.ValidForGame(game))
		}
	})

	t.Run("Schwarz+Schneider cannot be announced for suit games without Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.False(t, (GameModifierSchneiderAnnounced | GameModifierSchwarzAnnounced).ValidForGame(game))
		}
	})

	t.Run("Schwarz cannot be announced without Schneider", func(t *testing.T) {
		for _, game := range suitGames {
			assert.False(t, (GameModifierHand | GameModifierSchwarzAnnounced).ValidForGame(game))
		}
	})

	t.Run("Ouvert+Schwarz+Schneider cannot be announced for suit games without Hand", func(t *testing.T) {
		for _, game := range suitGames {
			assert.False(t, (GameModifierSchneiderAnnounced | GameModifierSchwarzAnnounced | GameModifierOuvert).ValidForGame(game))
		}
	})

	t.Run("Schneider can not be announced for null game", func(t *testing.T) {
		assert.False(t, GameModifierSchneiderAnnounced.ValidForGame(GameTypeNull))
	})

	t.Run("Schwarz can not be announced for null game", func(t *testing.T) {
		assert.False(t, GameModifierSchwarzAnnounced.ValidForGame(GameTypeNull))
	})

	t.Run("Ouvert can be announced for null game with Hand", func(t *testing.T) {
		assert.True(t, (GameModifierHand | GameModifierOuvert).ValidForGame(GameTypeNull))
	})

	t.Run("Ouvert can be announced for null game without Hand", func(t *testing.T) {
		assert.True(t, GameModifierOuvert.ValidForGame(GameTypeNull))
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

func TestCardEffectiveSuit(t *testing.T) {
	suitMap := map[Suit]EffectiveSuit{
		SuitBells:  EffectiveSuitBells,
		SuitHearts: EffectiveSuitHearts,
		SuitLeaves: EffectiveSuitLeaves,
		SuitAcorns: EffectiveSuitAcorns,
	}

	gameTypeMap := map[GameType]Suit{
		GameTypeBells:  SuitBells,
		GameTypeHearts: SuitHearts,
		GameTypeLeaves: SuitLeaves,
		GameTypeAcorns: SuitAcorns,
	}

	t.Run("grand leaves suits intact but maps jacks to trump", func(t *testing.T) {
		for _, c := range CardTypes {
			if c == CardJack {
				for _, suit := range Suits {
					assert.Equal(t, EffectiveSuitTrumps, c.As(suit).EffectiveSuit(GameTypeGrand))
				}
			} else {
				for suit, effective := range suitMap {
					assert.Equal(t, effective, c.As(suit).EffectiveSuit(GameTypeGrand))
				}
			}
		}
	})

	t.Run("suit game maps trump suit to trumps, leaves others intact", func(t *testing.T) {
		for gameType, trumpSuit := range gameTypeMap {
			for _, c := range CardTypes {
				if c == CardJack {
					for _, suit := range Suits {
						assert.Equal(t, EffectiveSuitTrumps, c.As(suit).EffectiveSuit(gameType))
					}
				} else {
					for suit, effective := range suitMap {
						if suit == trumpSuit {
							effective = EffectiveSuitTrumps
						}
						assert.Equal(t, effective, c.As(suit).EffectiveSuit(gameType))
					}
				}
			}
		}
	})

	t.Run("null maps all cards to their base effective suit", func(t *testing.T) {
		for _, c := range CardTypes {
			for suit, effective := range suitMap {
				assert.Equal(t, effective, c.As(suit).EffectiveSuit(GameTypeNull))
			}
		}
	})
}

func TestCardRelativePower(t *testing.T) {
	gameTypeMap := map[GameType]Suit{
		GameTypeBells:  SuitBells,
		GameTypeHearts: SuitHearts,
		GameTypeLeaves: SuitLeaves,
		GameTypeAcorns: SuitAcorns,
	}

	t.Run("grand jacks", func(t *testing.T) {
		prevSuit := Suits[0]
		for _, suit := range Suits[1:] {
			assert.Less(
				t,
				CardJack.As(prevSuit).RelativePower(GameTypeGrand),
				CardJack.As(suit).RelativePower(GameTypeGrand),
			)
			prevSuit = suit
		}
	})

	t.Run("grand non-jacks", func(t *testing.T) {
		cards := make([]CardType, 7)
		// card types without Jack
		copy(cards, CardTypes[:7])

		for _, suit := range Suits {
			prevCardType := cards[0]
			for _, cardType := range cards[1:] {
				assert.Less(
					t,
					prevCardType.As(suit).RelativePower(GameTypeGrand),
					cardType.As(suit).RelativePower(GameTypeGrand),
				)
				prevCardType = cardType
			}
		}
	})

	t.Run("null", func(t *testing.T) {
		nullCardOrder := []CardType{
			Card7, Card8, Card9, Card10, CardJack, CardQueen, CardKing,
			CardAce,
		}
		for _, suit := range Suits {
			prevCardType := nullCardOrder[0]
			for _, cardType := range nullCardOrder[1:] {
				assert.Less(
					t,
					prevCardType.As(suit).RelativePower(GameTypeNull),
					cardType.As(suit).RelativePower(GameTypeNull),
				)
				prevCardType = cardType
			}
		}
	})

	t.Run("suit game non-trumps", func(t *testing.T) {
		for _, gameType := range SuitGameTypes {
			for _, suit := range Suits {
				prevCardType := CardTypes[0]
				for _, cardType := range CardTypes[1:] {
					card := cardType.As(suit)
					if card.EffectiveSuit(gameType) == EffectiveSuitTrumps {
						// exclude trumps
						continue
					}
					assert.Less(
						t,
						prevCardType.As(suit).RelativePower(gameType),
						cardType.As(suit).RelativePower(gameType),
					)
					prevCardType = cardType
				}
			}
		}
	})

	t.Run("suit game trumps", func(t *testing.T) {
		for _, gameType := range SuitGameTypes {
			trumpSuit := gameTypeMap[gameType]
			expectedCardOrder := make([]Card, len(CardTypes)+3)
			for i, cardType := range CardTypes[:len(CardTypes)-1] {
				expectedCardOrder[i] = cardType.As(trumpSuit)
			}
			expectedCardOrder[len(CardTypes)-1] = CardJack.As(SuitBells)
			expectedCardOrder[len(CardTypes)] = CardJack.As(SuitHearts)
			expectedCardOrder[len(CardTypes)+1] = CardJack.As(SuitLeaves)
			expectedCardOrder[len(CardTypes)+2] = CardJack.As(SuitAcorns)

			prevCard := expectedCardOrder[0]
			for _, card := range expectedCardOrder[1:] {
				assert.Less(
					t,
					prevCard.RelativePower(gameType),
					card.RelativePower(gameType),
				)
				prevCard = card
			}
		}
	})
}

func TestTrickEvaluation(t *testing.T) {
	t.Run("effective suit is the suit of the first card", func(t *testing.T) {
		tr := Trick{Card7.As(SuitHearts), Card8.As(SuitBells), Card9.As(SuitAcorns)}
		assert.Equal(t, EffectiveSuitHearts, tr.EffectiveSuit(GameTypeGrand))
		assert.Equal(t, EffectiveSuitHearts, tr.EffectiveSuit(GameTypeBells))
		assert.Equal(t, EffectiveSuitTrumps, tr.EffectiveSuit(GameTypeHearts))
	})

	t.Run("card with highest relative power and matching suit wins", func(t *testing.T) {
		assert.Equal(
			t,
			0,
			Trick{Card9.As(SuitHearts), Card8.As(SuitHearts), Card7.As(SuitHearts)}.Taker(GameTypeGrand),
		)
		assert.Equal(
			t,
			1,
			Trick{Card7.As(SuitHearts), Card9.As(SuitHearts), Card8.As(SuitHearts)}.Taker(GameTypeGrand),
		)
		assert.Equal(
			t,
			2,
			Trick{Card7.As(SuitHearts), Card8.As(SuitHearts), Card9.As(SuitHearts)}.Taker(GameTypeGrand),
		)
	})

	t.Run("taker ignores mismatching suits", func(t *testing.T) {
		assert.Equal(
			t,
			2,
			Trick{Card7.As(SuitHearts), Card9.As(SuitBells), Card8.As(SuitHearts)}.Taker(GameTypeGrand),
		)
		assert.Equal(
			t,
			1,
			Trick{Card7.As(SuitHearts), Card8.As(SuitHearts), Card9.As(SuitBells)}.Taker(GameTypeGrand),
		)
	})

	t.Run("trumps win against other suits", func(t *testing.T) {
		assert.Equal(
			t,
			1,
			Trick{Card7.As(SuitHearts), CardJack.As(SuitBells), Card8.As(SuitHearts)}.Taker(GameTypeGrand),
		)
		assert.Equal(
			t,
			2,
			Trick{Card7.As(SuitHearts), Card8.As(SuitHearts), CardJack.As(SuitBells)}.Taker(GameTypeGrand),
		)
		assert.Equal(
			t,
			2,
			Trick{Card7.As(SuitHearts), Card8.As(SuitHearts), Card9.As(SuitBells)}.Taker(GameTypeBells),
		)
	})

	t.Run("trumps can win against other trumps", func(t *testing.T) {
		assert.Equal(
			t,
			1,
			Trick{Card7.As(SuitHearts), Card9.As(SuitBells), Card8.As(SuitBells)}.Taker(GameTypeBells),
		)
	})

	t.Run("null works as expected", func(t *testing.T) {
		assert.Equal(
			t,
			2,
			Trick{Card10.As(SuitHearts), CardJack.As(SuitBells), CardKing.As(SuitHearts)}.Taker(GameTypeNull),
		)
	})
}
