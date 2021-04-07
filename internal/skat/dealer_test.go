package skat

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPullUint8FromBytesExact(t *testing.T) {
	t.Run("returns byte if in range", func(t *testing.T) {
		buf := []byte{3}
		rd := bytes.NewReader(buf)
		v, err := PullUint8FromBytesExact(rd, 3)
		assert.Nil(t, err)
		assert.Equal(t, uint8(3), v)
	})

	t.Run("skips byte if out of range", func(t *testing.T) {
		buf := []byte{3}
		rd := bytes.NewReader(buf)
		_, err := PullUint8FromBytesExact(rd, 2)
		assert.NotNil(t, err)
		assert.Equal(t, err, io.EOF)
	})

	t.Run("uses mod logic to make use of more bytes", func(t *testing.T) {
		buf := []byte{7}
		rd := bytes.NewReader(buf)
		v, err := PullUint8FromBytesExact(rd, 3)
		assert.Nil(t, err)
		assert.Equal(t, uint8(3), v)
	})

	t.Run("skips byte and tries next if out of range", func(t *testing.T) {
		buf := []byte{3, 1}
		rd := bytes.NewReader(buf)
		v, err := PullUint8FromBytesExact(rd, 2)
		assert.Nil(t, err)
		assert.Equal(t, uint8(1), v)
	})
}

func TestDrawCards(t *testing.T) {
	t.Run("draws requested amount of cards", func(t *testing.T) {
		deck := NewCardDeck()
		remaining, drawn, err := DrawCards(deck, 3)
		assert.Nil(t, err)
		assert.Equal(t, 29, len(remaining))
		assert.Equal(t, 3, len(drawn))
	})

	t.Run("drawn and remaining are distinct", func(t *testing.T) {
		deck := NewCardDeck()
		remaining, drawn, err := DrawCards(deck, 16)
		assert.Nil(t, err)
		assert.Equal(t, 16, len(remaining))
		assert.Equal(t, 16, len(drawn))
		for _, remainingCard := range remaining {
			for _, drawnCard := range drawn {
				assert.NotEqual(t, remainingCard, drawnCard)
			}
		}
	})

	t.Run("no duplicate draws", func(t *testing.T) {
		deck := NewCardDeck()
		remaining, drawn, err := DrawCards(deck, 16)
		assert.Nil(t, err)
		assert.Equal(t, 16, len(remaining))
		assert.Equal(t, 16, len(drawn))
		for i, drawnCard1 := range drawn {
			for _, drawnCard2 := range drawn[i+1:] {
				assert.NotEqual(t, drawnCard1, drawnCard2)
			}
		}
	})
}

func TestShuffleDeckWithSeed(t *testing.T) {
	seed := []byte{23, 42}

	t.Run("deal a full hand", func(t *testing.T) {
		deck := NewCardDeck()
		err := ShuffleDeckWithSeed(seed, &deck)
		assert.Nil(t, err)
		var drawn []Card

		deck, drawn, err = DrawCards(deck, 3)
		assert.Nil(t, err)
		assert.Equal(t, 29, len(deck))
		assert.Equal(t, Card{Card8, SuitDiamonds}, drawn[0])
		assert.Equal(t, Card{CardKing, SuitClubs}, drawn[1])
		assert.Equal(t, Card{CardQueen, SuitHearts}, drawn[2])

		deck, drawn, err = DrawCards(deck, 3)
		assert.Nil(t, err)
		assert.Equal(t, 26, len(deck))
		assert.Equal(t, Card{Card7, SuitClubs}, drawn[0])
		assert.Equal(t, Card{Card8, SuitClubs}, drawn[1])
		assert.Equal(t, Card{Card9, SuitHearts}, drawn[2])

		deck, drawn, err = DrawCards(deck, 3)
		assert.Nil(t, err)
		assert.Equal(t, 23, len(deck))
		assert.Equal(t, Card{Card7, SuitDiamonds}, drawn[0])
		assert.Equal(t, Card{CardAce, SuitDiamonds}, drawn[1])
		assert.Equal(t, Card{CardJack, SuitSpades}, drawn[2])

		deck, drawn, err = DrawCards(deck, 2)
		assert.Nil(t, err)
		assert.Equal(t, 21, len(deck))
		assert.Equal(t, Card{CardKing, SuitHearts}, drawn[0])
		assert.Equal(t, Card{CardKing, SuitSpades}, drawn[1])

		deck, drawn, err = DrawCards(deck, 4)
		assert.Nil(t, err)
		assert.Equal(t, 17, len(deck))
		assert.Equal(t, Card{Card10, SuitClubs}, drawn[0])
		assert.Equal(t, Card{CardQueen, SuitClubs}, drawn[1])
		assert.Equal(t, Card{Card7, SuitSpades}, drawn[2])
		assert.Equal(t, Card{CardQueen, SuitSpades}, drawn[3])

		deck, drawn, err = DrawCards(deck, 4)
		assert.Nil(t, err)
		assert.Equal(t, 13, len(deck))
		assert.Equal(t, Card{CardJack, SuitClubs}, drawn[0])
		assert.Equal(t, Card{Card8, SuitHearts}, drawn[1])
		assert.Equal(t, Card{CardAce, SuitClubs}, drawn[2])
		assert.Equal(t, Card{Card10, SuitSpades}, drawn[3])

		deck, drawn, err = DrawCards(deck, 4)
		assert.Nil(t, err)
		assert.Equal(t, 9, len(deck))
		assert.Equal(t, Card{Card7, SuitHearts}, drawn[0])
		assert.Equal(t, Card{Card10, SuitDiamonds}, drawn[1])
		assert.Equal(t, Card{CardJack, SuitDiamonds}, drawn[2])
		assert.Equal(t, Card{CardQueen, SuitDiamonds}, drawn[3])

		deck, drawn, err = DrawCards(deck, 3)
		assert.Nil(t, err)
		assert.Equal(t, 6, len(deck))
		assert.Equal(t, Card{Card10, SuitHearts}, drawn[0])
		assert.Equal(t, Card{CardAce, SuitHearts}, drawn[1])
		assert.Equal(t, Card{CardJack, SuitHearts}, drawn[2])

		deck, drawn, err = DrawCards(deck, 3)
		assert.Nil(t, err)
		assert.Equal(t, 3, len(deck))
		assert.Equal(t, Card{Card9, SuitSpades}, drawn[0])
		assert.Equal(t, Card{CardAce, SuitSpades}, drawn[1])
		assert.Equal(t, Card{Card8, SuitSpades}, drawn[2])

		deck, drawn, err = DrawCards(deck, 3)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(deck))
		assert.Equal(t, Card{CardKing, SuitDiamonds}, drawn[0])
		assert.Equal(t, Card{Card9, SuitDiamonds}, drawn[1])
		assert.Equal(t, Card{Card9, SuitClubs}, drawn[2])
	})
}
