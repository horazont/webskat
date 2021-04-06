package skat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testGetOnTheEdgeGame(t *testing.T) [3]CardSet {
	// 60 points on one stack, ~30 points distributed among the others
	// stack 1: 11 11 4 2 2 = 30
	// stack 2: 10 10 10 10 11 3 3 3 = 60
	// stack 3: 11 4 4 4 3 2 2 = 30

	deck := NewCardDeck()
	player1 := make(CardSet, 0)
	player2 := make(CardSet, 0)
	player3 := make(CardSet, 0)

	assert.Equal(t, 32, len(deck))

	assert.Nil(t, testTransferCard(&deck, &player1, CardAce.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &player1, CardAce.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &player1, CardKing.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &player1, CardJack.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &player1, CardJack.As(SuitHearts)))

	assert.Equal(t, 30, player1.Value())

	assert.Nil(t, testTransferCard(&deck, &player2, Card10.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &player2, Card10.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &player2, Card10.As(SuitSpades)))
	assert.Nil(t, testTransferCard(&deck, &player2, Card10.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &player2, CardAce.As(SuitSpades)))
	assert.Nil(t, testTransferCard(&deck, &player2, CardQueen.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &player2, CardQueen.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &player2, CardQueen.As(SuitSpades)))

	assert.Equal(t, 60, player2.Value())

	assert.Nil(t, testTransferCard(&deck, &player3, CardAce.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardKing.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardKing.As(SuitSpades)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardKing.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardQueen.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardJack.As(SuitSpades)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardJack.As(SuitClubs)))

	assert.Equal(t, 30, player3.Value())

	result := [3]CardSet{player1, player2, player3}
	for i, leftover := range deck {
		var err error
		assert.Equal(t, 0, leftover.Value())
		result[i%3], err = result[i%3].Push(leftover)
		assert.Nil(t, err)
	}

	return result
}

func testSchwarzGame(t *testing.T) [3]CardSet {
	player1 := make(CardSet, 0)
	player2 := NewCardDeck()
	player3 := make(CardSet, 0)

	return [3]CardSet{player1, player2, player3}
}

func testPseudoSchwarzGame(t *testing.T) [3]CardSet {
	player1 := make(CardSet, 0)
	player2 := NewCardDeck()
	player3 := make(CardSet, 0)

	assert.Nil(t, testTransferCard(&player2, &player1, Card7.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&player2, &player1, Card8.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&player2, &player1, Card9.As(SuitHearts)))

	return [3]CardSet{player1, player2, player3}
}

func testBalancedGame(t *testing.T) [3]CardSet {
	// 60 points on one stack, ~30 points distributed among the others
	// stack 1: 11 2 2 = 15
	// stack 2: 10 10 10 10 11 11 3 3 3 = 71
	// stack 3: 11 4 4 4 4 3 2 2 = 34

	deck := NewCardDeck()
	player1 := make(CardSet, 0)
	player2 := make(CardSet, 0)
	player3 := make(CardSet, 0)

	assert.Equal(t, 32, len(deck))

	assert.Nil(t, testTransferCard(&deck, &player1, CardAce.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &player1, CardJack.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &player1, CardJack.As(SuitHearts)))

	assert.Equal(t, 15, player1.Value())

	assert.Nil(t, testTransferCard(&deck, &player2, Card10.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &player2, Card10.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &player2, Card10.As(SuitSpades)))
	assert.Nil(t, testTransferCard(&deck, &player2, Card10.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &player2, CardAce.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &player2, CardAce.As(SuitSpades)))
	assert.Nil(t, testTransferCard(&deck, &player2, CardQueen.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &player2, CardQueen.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &player2, CardQueen.As(SuitSpades)))

	assert.Equal(t, 71, player2.Value())

	assert.Nil(t, testTransferCard(&deck, &player3, CardAce.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardKing.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardKing.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardKing.As(SuitSpades)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardKing.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardQueen.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardJack.As(SuitSpades)))
	assert.Nil(t, testTransferCard(&deck, &player3, CardJack.As(SuitClubs)))

	assert.Equal(t, 34, player3.Value())

	result := [3]CardSet{player1, player2, player3}
	for i, leftover := range deck {
		var err error
		assert.Equal(t, 0, leftover.Value())
		result[i%3], err = result[i%3].Push(leftover)
		assert.Nil(t, err)
	}

	return result
}

func TestGetMatadorsJackStrength(t *testing.T) {
	baseHand := CardSet{
		CardJack.As(SuitClubs),
		CardJack.As(SuitSpades),
		CardJack.As(SuitHearts),
		CardJack.As(SuitDiamonds),
		CardAce.As(SuitHearts),
		CardAce.As(SuitSpades),
		Card10.As(SuitSpades),
		CardKing.As(SuitSpades),
		CardQueen.As(SuitSpades),
		Card9.As(SuitSpades),
		Card8.As(SuitSpades),
		Card7.As(SuitDiamonds),
		Card7.As(SuitHearts),
		Card7.As(SuitSpades),
	}

	t.Run("suit game with: full strength", func(t *testing.T) {
		hand := baseHand.Copy()
		assert.Equal(t, 11, hand.GetMatadorsJackStrength(GameTypeSpades))
	})

	t.Run("suit game without: full strength", func(t *testing.T) {
		hand := baseHand.Copy()
		hand, _ = hand.Pop(CardJack.As(SuitClubs))
		hand, _ = hand.Pop(CardJack.As(SuitSpades))
		hand, _ = hand.Pop(CardJack.As(SuitHearts))
		hand, _ = hand.Pop(CardJack.As(SuitDiamonds))
		assert.Equal(t, 11, hand.GetMatadorsJackStrength(GameTypeClubs))
	})

	t.Run("suit game with: gaps", func(t *testing.T) {
		hand := baseHand.Copy()
		hand, _ = hand.Pop(CardJack.As(SuitHearts))
		assert.Equal(t, 2, hand.GetMatadorsJackStrength(GameTypeSpades))

		hand = baseHand.Copy()
		assert.Equal(t, 5, hand.GetMatadorsJackStrength(GameTypeHearts))

		hand = baseHand.Copy()
		hand, _ = hand.Pop(CardAce.As(SuitSpades))
		assert.Equal(t, 4, hand.GetMatadorsJackStrength(GameTypeSpades))
	})

	t.Run("suit game without: gaps", func(t *testing.T) {
		hand := baseHand.Copy()
		hand, _ = hand.Pop(CardJack.As(SuitClubs))
		hand, _ = hand.Pop(CardJack.As(SuitSpades))
		assert.Equal(t, 2, hand.GetMatadorsJackStrength(GameTypeSpades))

		hand = baseHand.Copy()
		hand, _ = hand.Pop(CardJack.As(SuitClubs))
		hand, _ = hand.Pop(CardJack.As(SuitSpades))
		hand, _ = hand.Pop(CardJack.As(SuitHearts))
		hand, _ = hand.Pop(CardJack.As(SuitDiamonds))
		assert.Equal(t, 4, hand.GetMatadorsJackStrength(GameTypeSpades))
	})

	t.Run("grand game with: full strength", func(t *testing.T) {
		hand := baseHand.Copy()
		assert.Equal(t, 4, hand.GetMatadorsJackStrength(GameTypeGrand))
	})

	t.Run("grand game without: full strength", func(t *testing.T) {
		hand := baseHand.Copy()
		hand, _ = hand.Pop(CardJack.As(SuitClubs))
		hand, _ = hand.Pop(CardJack.As(SuitSpades))
		hand, _ = hand.Pop(CardJack.As(SuitHearts))
		hand, _ = hand.Pop(CardJack.As(SuitDiamonds))
		assert.Equal(t, 4, hand.GetMatadorsJackStrength(GameTypeGrand))
	})

	t.Run("grand game with: gaps", func(t *testing.T) {
		hand := baseHand.Copy()
		hand, _ = hand.Pop(CardJack.As(SuitSpades))
		assert.Equal(t, 1, hand.GetMatadorsJackStrength(GameTypeGrand))
	})

	t.Run("grand game without: gaps", func(t *testing.T) {
		hand := baseHand.Copy()
		hand, _ = hand.Pop(CardJack.As(SuitClubs))
		hand, _ = hand.Pop(CardJack.As(SuitHearts))
		assert.Equal(t, 1, hand.GetMatadorsJackStrength(GameTypeGrand))
	})

	t.Run("null game", func(t *testing.T) {
		hand := baseHand.Copy()
		assert.Equal(t, 0, hand.GetMatadorsJackStrength(GameTypeNull))
	})
}

func TestCalculateGameValue(t *testing.T) {
	factorOneHand := CardSet{
		CardJack.As(SuitClubs),
		CardJack.As(SuitHearts),
	}

	baseHand := CardSet{
		CardJack.As(SuitClubs),
		CardJack.As(SuitSpades),
		CardJack.As(SuitHearts),
		CardJack.As(SuitDiamonds),
		CardAce.As(SuitHearts),
		CardAce.As(SuitSpades),
		Card10.As(SuitSpades),
		CardKing.As(SuitSpades),
		CardQueen.As(SuitSpades),
		Card9.As(SuitSpades),
		Card8.As(SuitSpades),
		Card7.As(SuitDiamonds),
		Card7.As(SuitHearts),
		Card7.As(SuitSpades),
	}

	nonSpecialGameTypes := append([]GameType{
		GameTypeGrand,
	}, SuitGameTypes...)

	t.Run("null game", func(t *testing.T) {
		base, factor := CalculateGameValue(nil, GameTypeNull, NoGameModifiers)
		assert.Equal(t, 23, base)
		assert.Equal(t, 1, factor)
	})

	t.Run("null+hand game", func(t *testing.T) {
		base, factor := CalculateGameValue(nil, GameTypeNull, GameModifierHand)
		assert.Equal(t, 35, base)
		assert.Equal(t, 1, factor)
	})

	t.Run("null+ouvert game", func(t *testing.T) {
		base, factor := CalculateGameValue(nil, GameTypeNull, GameModifierOuvert)
		assert.Equal(t, 46, base)
		assert.Equal(t, 1, factor)
	})

	t.Run("null+hand+ouvert game", func(t *testing.T) {
		base, factor := CalculateGameValue(nil, GameTypeNull, GameModifierHand|GameModifierOuvert)
		assert.Equal(t, 59, base)
		assert.Equal(t, 1, factor)
	})

	t.Run("bells game base value", func(t *testing.T) {
		base, factor := CalculateGameValue(factorOneHand, GameTypeDiamonds, NoGameModifiers)
		assert.Equal(t, 9, base)
		assert.Equal(t, 2, factor)
	})

	t.Run("hearts game base value", func(t *testing.T) {
		base, factor := CalculateGameValue(factorOneHand, GameTypeHearts, NoGameModifiers)
		assert.Equal(t, 10, base)
		assert.Equal(t, 2, factor)
	})

	t.Run("leaves game base value", func(t *testing.T) {
		base, factor := CalculateGameValue(factorOneHand, GameTypeSpades, NoGameModifiers)
		assert.Equal(t, 11, base)
		assert.Equal(t, 2, factor)
	})

	t.Run("acorns game base value", func(t *testing.T) {
		base, factor := CalculateGameValue(factorOneHand, GameTypeClubs, NoGameModifiers)
		assert.Equal(t, 12, base)
		assert.Equal(t, 2, factor)
	})

	t.Run("suit game matadors jack strength used", func(t *testing.T) {
		var hand CardSet
		var factor int

		hand = baseHand.Copy()
		_, factor = CalculateGameValue(hand, GameTypeSpades, NoGameModifiers)
		assert.Equal(t, 12, factor)

		hand = baseHand.Copy()
		hand, _ = hand.Pop(CardJack.As(SuitHearts))
		_, factor = CalculateGameValue(hand, GameTypeSpades, NoGameModifiers)
		assert.Equal(t, 3, factor)

		hand = baseHand.Copy()
		hand, _ = hand.Pop(CardJack.As(SuitClubs))
		hand, _ = hand.Pop(CardJack.As(SuitSpades))
		hand, _ = hand.Pop(CardJack.As(SuitHearts))
		_, factor = CalculateGameValue(hand, GameTypeSpades, NoGameModifiers)
		assert.Equal(t, 4, factor)
	})

	t.Run("suit game modifiers: hand adds one", func(t *testing.T) {
		for _, gameType := range nonSpecialGameTypes {
			var factor int
			_, factor = CalculateGameValue(factorOneHand, gameType, GameModifierHand)
			assert.Equal(t, 3, factor)
		}
	})

	t.Run("suit game modifiers: schneider adds one", func(t *testing.T) {
		for _, gameType := range nonSpecialGameTypes {
			var factor int
			_, factor = CalculateGameValue(factorOneHand, gameType, GameModifierSchneider)
			assert.Equal(t, 3, factor)
		}
	})

	t.Run("suit game modifiers: schwarz adds two", func(t *testing.T) {
		for _, gameType := range nonSpecialGameTypes {
			var factor int
			_, factor = CalculateGameValue(factorOneHand, gameType, GameModifierSchwarz.Normalized())
			assert.Equal(t, 4, factor)
		}
	})

	t.Run("suit game modifiers: schneider announced adds one", func(t *testing.T) {
		for _, gameType := range nonSpecialGameTypes {
			var factor int
			_, factor = CalculateGameValue(factorOneHand, gameType, GameModifierHand|GameModifierSchneiderAnnounced)
			assert.Equal(t, 4, factor)
		}
	})

	t.Run("suit game modifiers: schwarz announced adds two", func(t *testing.T) {
		for _, gameType := range nonSpecialGameTypes {
			var factor int
			_, factor = CalculateGameValue(factorOneHand, gameType, (GameModifierHand | GameModifierSchwarzAnnounced).Normalized())
			assert.Equal(t, 5, factor)
		}
	})

	t.Run("suit game modifiers: ouvert adds one", func(t *testing.T) {
		for _, gameType := range nonSpecialGameTypes {
			var factor int
			_, factor = CalculateGameValue(factorOneHand, gameType, GameModifierOuvert.Normalized())
			assert.Equal(t, 3, factor)
		}
	})

	t.Run("suit game modifiers: all the modifiers", func(t *testing.T) {
		for _, gameType := range nonSpecialGameTypes {
			var factor int
			_, factor = CalculateGameValue(factorOneHand, gameType, (GameModifierHand | GameModifierSchwarz | GameModifierSchwarzAnnounced | GameModifierOuvert).Normalized())
			assert.Equal(t, 8, factor)
		}
	})
}

func TestEvaluateWonCards(t *testing.T) {
	t.Run("evaluates team scores", func(t *testing.T) {
		cards := testGetOnTheEdgeGame(t)
		_, declarer, defender := EvaluateWonCards(cards, 0)
		assert.Equal(t, 30, declarer)
		assert.Equal(t, 90, defender)

		_, declarer, defender = EvaluateWonCards(cards, 1)
		assert.Equal(t, 60, declarer)
		assert.Equal(t, 60, defender)

		_, declarer, defender = EvaluateWonCards(cards, 2)
		assert.Equal(t, 30, declarer)
		assert.Equal(t, 90, defender)

		cards = testBalancedGame(t)
		_, declarer, defender = EvaluateWonCards(cards, 0)
		assert.Equal(t, 15, declarer)
		assert.Equal(t, 120-15, defender)

		_, declarer, defender = EvaluateWonCards(cards, 1)
		assert.Equal(t, 71, declarer)
		assert.Equal(t, 49, defender)

		_, declarer, defender = EvaluateWonCards(cards, 2)
		assert.Equal(t, 34, declarer)
		assert.Equal(t, 120-34, defender)

		cards = testSchwarzGame(t)
		_, declarer, defender = EvaluateWonCards(cards, 0)
		assert.Equal(t, 0, declarer)
		assert.Equal(t, 120, defender)

		_, declarer, defender = EvaluateWonCards(cards, 1)
		assert.Equal(t, 120, declarer)
		assert.Equal(t, 0, defender)

		_, declarer, defender = EvaluateWonCards(cards, 2)
		assert.Equal(t, 0, declarer)
		assert.Equal(t, 120, defender)

		cards = testPseudoSchwarzGame(t)
		_, declarer, defender = EvaluateWonCards(cards, 0)
		assert.Equal(t, 0, declarer)
		assert.Equal(t, 120, defender)

		_, declarer, defender = EvaluateWonCards(cards, 1)
		assert.Equal(t, 120, declarer)
		assert.Equal(t, 0, defender)

		_, declarer, defender = EvaluateWonCards(cards, 2)
		assert.Equal(t, 0, declarer)
		assert.Equal(t, 120, defender)
	})

	t.Run("detect Schneider modifier correctly", func(t *testing.T) {
		var modifiers GameModifier
		cards := testGetOnTheEdgeGame(t)

		modifiers, _, _ = EvaluateWonCards(cards, 0)
		assert.Equal(t, GameModifierSchneider, modifiers)

		modifiers, _, _ = EvaluateWonCards(cards, 1)
		assert.Equal(t, NoGameModifiers, modifiers)

		modifiers, _, _ = EvaluateWonCards(cards, 2)
		assert.Equal(t, GameModifierSchneider, modifiers)
	})

	t.Run("detect Schwarz modifier correctly", func(t *testing.T) {
		var modifiers GameModifier
		cards := testSchwarzGame(t)

		modifiers, _, _ = EvaluateWonCards(cards, 0)
		assert.Equal(t, GameModifierSchwarz.Normalized(), modifiers)

		modifiers, _, _ = EvaluateWonCards(cards, 1)
		assert.Equal(t, GameModifierSchwarz.Normalized(), modifiers)

		modifiers, _, _ = EvaluateWonCards(cards, 2)
		assert.Equal(t, GameModifierSchwarz.Normalized(), modifiers)

		cards = testPseudoSchwarzGame(t)

		modifiers, _, _ = EvaluateWonCards(cards, 0)
		assert.Equal(t, GameModifierSchneider, modifiers)

		modifiers, _, _ = EvaluateWonCards(cards, 1)
		assert.Equal(t, GameModifierSchneider, modifiers)

		modifiers, _, _ = EvaluateWonCards(cards, 2)
		assert.Equal(t, GameModifierSchwarz.Normalized(), modifiers)
	})
}

func TestEvaluateGame(t *testing.T) {
	t.Run("non-overbid won by declarer", func(t *testing.T) {
		won, gameValue, reason := EvaluateGame(
			9,
			2,
			120,
			18,
			0,
			NoGameModifiers,
		)
		assert.True(t, won)
		assert.Equal(t, 18, gameValue)
		assert.Equal(t, "", reason)
	})

	t.Run("non-overbid lost by declarer", func(t *testing.T) {
		won, gameValue, reason := EvaluateGame(
			9,
			2,
			60,
			18,
			0,
			NoGameModifiers,
		)
		assert.False(t, won)
		assert.Equal(t, 18, gameValue)
		assert.Equal(t, LossReasonNotEnoughPoints, reason)
	})

	t.Run("overbid rounds up game value", func(t *testing.T) {
		won, gameValue, reason := EvaluateGame(
			9,
			2,
			70,
			20,
			0,
			NoGameModifiers,
		)
		assert.False(t, won)
		assert.Equal(t, 27, gameValue)
		assert.Equal(t, LossReasonOverbid, reason)

		won, gameValue, reason = EvaluateGame(
			9,
			2,
			70,
			18,
			0,
			NoGameModifiers,
		)
		assert.Equal(t, 18, gameValue)
	})

	t.Run("overbid takes precedence over point loss", func(t *testing.T) {
		won, _, reason := EvaluateGame(
			9,
			2,
			30,
			20,
			0,
			NoGameModifiers,
		)
		assert.False(t, won)
		assert.Equal(t, LossReasonOverbid, reason)
	})

	t.Run("overbid takes precedence over modifier loss", func(t *testing.T) {
		won, _, reason := EvaluateGame(
			9,
			2,
			30,
			20,
			0,
			GameModifierSchneiderAnnounced,
		)
		assert.False(t, won)
		assert.Equal(t, LossReasonOverbid, reason)
	})

	t.Run("announced schneider but did not make it", func(t *testing.T) {
		won, gameValue, reason := EvaluateGame(
			9,
			2,
			120,
			18,
			0,
			GameModifierSchneiderAnnounced,
		)
		assert.False(t, won)
		assert.Equal(t, 18, gameValue)
		assert.Equal(t, LossReasonNoSchneider, reason)
	})

	t.Run("announced schwarz but did not make it", func(t *testing.T) {
		won, gameValue, reason := EvaluateGame(
			9,
			2,
			120,
			18,
			0,
			GameModifierSchwarzAnnounced.Normalized(),
		)
		assert.False(t, won)
		assert.Equal(t, 18, gameValue)
		assert.Equal(t, LossReasonNoSchwarz, reason)

		won, gameValue, reason = EvaluateGame(
			9,
			2,
			120,
			18,
			0,
			GameModifierSchwarzAnnounced.Normalized().With(GameModifierSchneider),
		)
		assert.False(t, won)
		assert.Equal(t, 18, gameValue)
		assert.Equal(t, LossReasonNoSchwarz, reason)
	})

	t.Run("null game with any cards", func(t *testing.T) {
		won, gameValue, reason := EvaluateGame(
			35,
			1,
			0,
			18,
			GameTypeNull,
			NoGameModifiers,
		)
		assert.False(t, won)
		assert.Equal(t, 35, gameValue)
		assert.Equal(t, LossReasonNotNull, reason)

		won, gameValue, reason = EvaluateGame(
			35,
			1,
			120,
			18,
			GameTypeNull,
			GameModifierSchwarz.Normalized(),
		)
		assert.False(t, won)
		assert.Equal(t, 35, gameValue)
		assert.Equal(t, LossReasonNotNull, reason)
	})

	t.Run("won null game", func(t *testing.T) {
		won, gameValue, reason := EvaluateGame(
			35,
			1,
			0,
			18,
			GameTypeNull,
			GameModifierSchwarz.Normalized(),
		)
		assert.True(t, won)
		assert.Equal(t, 35, gameValue)
		assert.Equal(t, "", reason)
	})
}
