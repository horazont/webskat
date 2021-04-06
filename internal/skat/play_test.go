package skat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type playingStateTest struct {
	handsBuf [3]CardSet
	s        *PlayingState
}

func testTransferCard(fromSet *CardSet, toSet *CardSet, c Card) error {
	newFromSet, err := fromSet.Pop(c)
	if err != nil {
		return err
	}
	newToSet, err := toSet.Push(c)
	if err != nil {
		return err
	}
	*fromSet = newFromSet
	*toSet = newToSet
	return nil
}

func testPlayingState(t *testing.T) *playingStateTest {
	deck := NewCardDeck()
	forehand := CardSet{}
	middlehand := CardSet{}
	rearhand := CardSet{}
	pushed := CardSet{}

	assert.Equal(t, 32, len(deck))

	assert.Nil(t, testTransferCard(&deck, &pushed, Card7.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &pushed, Card8.As(SuitDiamonds)))

	assert.Equal(t, 30, len(deck))

	assert.Nil(t, testTransferCard(&deck, &forehand, CardJack.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &forehand, CardJack.As(SuitSpades)))
	assert.Nil(t, testTransferCard(&deck, &forehand, CardAce.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &forehand, Card10.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &forehand, CardKing.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &forehand, CardQueen.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &forehand, Card7.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &forehand, CardAce.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &forehand, Card10.As(SuitSpades)))
	assert.Nil(t, testTransferCard(&deck, &forehand, Card8.As(SuitClubs)))

	assert.Equal(t, 20, len(deck))

	assert.Nil(t, testTransferCard(&deck, &middlehand, CardJack.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &middlehand, CardAce.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &middlehand, Card10.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &middlehand, CardQueen.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &middlehand, Card9.As(SuitDiamonds)))
	assert.Nil(t, testTransferCard(&deck, &middlehand, CardQueen.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &middlehand, Card7.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &middlehand, Card8.As(SuitHearts)))
	assert.Nil(t, testTransferCard(&deck, &middlehand, Card9.As(SuitClubs)))
	assert.Nil(t, testTransferCard(&deck, &middlehand, CardKing.As(SuitClubs)))

	assert.Equal(t, 10, len(deck))

	for _, card := range deck.Copy() {
		assert.Nil(t, testTransferCard(&deck, &rearhand, card))
	}

	assert.Equal(t, 0, len(deck))

	result := &playingStateTest{
		handsBuf: [3]CardSet{
			forehand,
			middlehand,
			rearhand,
		},
	}
	result.s = NewPlayingState(
		PlayerInitialForehand,
		GameTypeHearts,
		[3]*CardSet{
			&result.handsBuf[0],
			&result.handsBuf[1],
			&result.handsBuf[2],
		},
		pushed,
	)
	return result
}

func TestPlayingStateInit(t *testing.T) {
	t.Run("last trick is undefined", func(t *testing.T) {
		ts := testPlayingState(t)
		_, player := ts.s.GetLastTrick()
		assert.Equal(t, PlayerNone, player)
	})

	t.Run("table is empty", func(t *testing.T) {
		ts := testPlayingState(t)
		cards := ts.s.GetTable()
		assert.Equal(t, 0, len(cards))
	})

	t.Run("initial forehand is current player", func(t *testing.T) {
		ts := testPlayingState(t)
		assert.Equal(t, PlayerInitialForehand, ts.s.GetCurrentPlayer())
	})

	t.Run("pushed cards are added to player won cards", func(t *testing.T) {
		ts := testPlayingState(t)
		wonCards := ts.s.GetWonCards(PlayerInitialForehand)
		assert.Equal(t, CardSet{Card{Card7, SuitDiamonds}, Card{Card8, SuitDiamonds}}, wonCards)
	})

	t.Run("pushed cards are not added to other players won cards", func(t *testing.T) {
		ts := testPlayingState(t)
		wonCards := ts.s.GetWonCards(PlayerInitialMiddlehand)
		assert.Equal(t, CardSet{}, wonCards)

		wonCards = ts.s.GetWonCards(PlayerInitialRearhand)
		assert.Equal(t, CardSet{}, wonCards)
	})
}

func TestPlayTrick(t *testing.T) {
	t.Run("playing a card removes it from the hand", func(t *testing.T) {
		ts := testPlayingState(t)
		card := CardJack.As(SuitSpades)
		assert.Nil(t, ts.s.Play(PlayerInitialForehand, card))
		assert.False(t, ts.s.GetHand(PlayerInitialForehand).Contains(card))
	})

	t.Run("playing a card adds it to the table", func(t *testing.T) {
		ts := testPlayingState(t)
		card := CardJack.As(SuitSpades)
		assert.Nil(t, ts.s.Play(PlayerInitialForehand, card))
		assert.True(t, ts.s.GetTable().Contains(card))
	})

	t.Run("playing a card advances current player", func(t *testing.T) {
		ts := testPlayingState(t)
		card := CardJack.As(SuitSpades)
		assert.Nil(t, ts.s.Play(PlayerInitialForehand, card))
		assert.Equal(t, PlayerInitialMiddlehand, ts.s.GetCurrentPlayer())
	})

	t.Run("reject playing card not in hand", func(t *testing.T) {
		ts := testPlayingState(t)
		card := CardJack.As(SuitDiamonds)
		assert.Equal(t, ErrCardNotPresent, ts.s.Play(PlayerInitialForehand, card))
	})

	t.Run("reject playing card by wrong player", func(t *testing.T) {
		ts := testPlayingState(t)
		card := CardJack.As(SuitDiamonds)
		assert.Equal(t, ErrNotYourTurn, ts.s.Play(PlayerInitialMiddlehand, card))
	})

	t.Run("play complete trick", func(t *testing.T) {
		ts := testPlayingState(t)
		card1 := CardJack.As(SuitSpades)
		card2 := Card8.As(SuitHearts)
		card3 := Card9.As(SuitHearts)
		assert.Nil(t, ts.s.Play(PlayerInitialForehand, card1))
		assert.Nil(t, ts.s.Play(PlayerInitialMiddlehand, card2))
		assert.Nil(t, ts.s.Play(PlayerInitialRearhand, card3))
		assert.Equal(t, CardSet{}, ts.s.GetTable())
		trick, taker := ts.s.GetLastTrick()
		assert.Equal(t, Trick{card1, card2, card3}, trick)
		assert.Equal(t, PlayerInitialForehand, taker)
		assert.Equal(t, trick.AsCardSet(), ts.s.GetWonCards(PlayerInitialForehand)[2:])
		assert.Equal(t, CardSet{}, ts.s.GetWonCards(PlayerInitialMiddlehand))
		assert.Equal(t, CardSet{}, ts.s.GetWonCards(PlayerInitialRearhand))
		assert.Equal(t, PlayerInitialForehand, ts.s.GetCurrentPlayer())
	})

	t.Run("play complete with change of forehandship", func(t *testing.T) {
		ts := testPlayingState(t)
		card1 := Card8.As(SuitClubs)
		card2 := CardKing.As(SuitClubs)
		card3 := Card10.As(SuitClubs)
		assert.Nil(t, ts.s.Play(PlayerInitialForehand, card1))
		assert.Nil(t, ts.s.Play(PlayerInitialMiddlehand, card2))
		assert.Nil(t, ts.s.Play(PlayerInitialRearhand, card3))
		assert.Equal(t, CardSet{}, ts.s.GetTable())
		trick, taker := ts.s.GetLastTrick()
		assert.Equal(t, Trick{card1, card2, card3}, trick)
		assert.Equal(t, PlayerInitialRearhand, taker)
		assert.Equal(t, trick.AsCardSet(), ts.s.GetWonCards(PlayerInitialRearhand))
		assert.Equal(t, CardSet{}, ts.s.GetWonCards(PlayerInitialMiddlehand))
		assert.Equal(t, CardSet{}, ts.s.GetWonCards(PlayerInitialForehand)[2:])
		assert.Equal(t, PlayerInitialRearhand, ts.s.GetCurrentPlayer())
	})

	t.Run("play two tricks with forehandship change", func(t *testing.T) {
		ts := testPlayingState(t)
		card1 := Card7.As(SuitHearts)
		card2 := Card8.As(SuitHearts)
		card3 := Card9.As(SuitHearts)
		assert.Nil(t, ts.s.Play(PlayerInitialForehand, card1))
		assert.Nil(t, ts.s.Play(PlayerInitialMiddlehand, card2))
		assert.Nil(t, ts.s.Play(PlayerInitialRearhand, card3))
		assert.Equal(t, CardSet{}, ts.s.GetTable())
		trick, taker := ts.s.GetLastTrick()
		assert.Equal(t, Trick{card1, card2, card3}, trick)
		assert.Equal(t, PlayerInitialRearhand, taker)
		assert.Equal(t, trick.AsCardSet(), ts.s.GetWonCards(PlayerInitialRearhand))
		assert.Equal(t, CardSet{}, ts.s.GetWonCards(PlayerInitialMiddlehand))
		assert.Equal(t, CardSet{}, ts.s.GetWonCards(PlayerInitialForehand)[2:])
		assert.Equal(t, PlayerInitialRearhand, ts.s.GetCurrentPlayer())

		card1 = Card10.As(SuitClubs)
		card2 = Card8.As(SuitClubs)
		card3 = CardAce.As(SuitClubs)
		assert.Nil(t, ts.s.Play(PlayerInitialRearhand, card1))
		assert.Nil(t, ts.s.Play(PlayerInitialForehand, card2))
		assert.Nil(t, ts.s.Play(PlayerInitialMiddlehand, card3))
		assert.Equal(t, CardSet{}, ts.s.GetTable())
		trick, taker = ts.s.GetLastTrick()
		assert.Equal(t, Trick{card1, card2, card3}, trick)
		assert.Equal(t, PlayerInitialMiddlehand, taker)
		assert.Equal(t, trick.AsCardSet(), ts.s.GetWonCards(PlayerInitialMiddlehand))
		assert.Equal(t, CardSet{}, ts.s.GetWonCards(PlayerInitialForehand)[2:])
		assert.Equal(t, CardSet{}, ts.s.GetWonCards(PlayerInitialRearhand)[3:])
		assert.Equal(t, PlayerInitialMiddlehand, ts.s.GetCurrentPlayer())
	})

	t.Run("reject missing suit following", func(t *testing.T) {
		ts := testPlayingState(t)
		card1 := CardJack.As(SuitSpades)
		card2 := Card9.As(SuitDiamonds)
		assert.Nil(t, ts.s.Play(PlayerInitialForehand, card1))
		assert.Equal(t, ErrMustFollowSuit, ts.s.Play(PlayerInitialMiddlehand, card2))
	})

	t.Run("allow diverging from suit if impossible", func(t *testing.T) {
		ts := testPlayingState(t)
		card1 := Card10.As(SuitSpades)
		card2 := Card10.As(SuitDiamonds)
		card3 := CardAce.As(SuitSpades)
		assert.Nil(t, ts.s.Play(PlayerInitialForehand, card1))
		assert.Nil(t, ts.s.Play(PlayerInitialMiddlehand, card2))
		assert.Nil(t, ts.s.Play(PlayerInitialRearhand, card3))
		assert.Equal(t, CardSet{}, ts.s.GetTable())
		trick, taker := ts.s.GetLastTrick()
		assert.Equal(t, Trick{card1, card2, card3}, trick)
		assert.Equal(t, PlayerInitialRearhand, taker)
		assert.Equal(t, trick.AsCardSet(), ts.s.GetWonCards(PlayerInitialRearhand))
		assert.Equal(t, CardSet{}, ts.s.GetWonCards(PlayerInitialMiddlehand))
		assert.Equal(t, CardSet{}, ts.s.GetWonCards(PlayerInitialForehand)[2:])
		assert.Equal(t, PlayerInitialRearhand, ts.s.GetCurrentPlayer())
	})
}
