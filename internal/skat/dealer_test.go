package skat

import (
	"bytes"
	"crypto/rand"
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

	t.Run("skips byte and tries next if out of range", func(t *testing.T) {
		buf := []byte{3, 1}
		rd := bytes.NewReader(buf)
		v, err := PullUint8FromBytesExact(rd, 2)
		assert.Nil(t, err)
		assert.Equal(t, uint8(1), v)
	})
}

func TestDeterministicDealerNonDeterministic(t *testing.T) {
	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	assert.Nil(t, err)

	t.Run("draws requested amount of cards", func(t *testing.T) {
		deck := NewCardDeck()
		d, err := NewDealer(seed)
		assert.Nil(t, err)
		remaining, drawn, err := d.DrawCards(deck, 3)
		assert.Nil(t, err)
		assert.Equal(t, 29, len(remaining))
		assert.Equal(t, 3, len(drawn))
	})

	t.Run("drawn and remaining are distinct", func(t *testing.T) {
		deck := NewCardDeck()
		d, err := NewDealer(seed)
		assert.Nil(t, err)
		remaining, drawn, err := d.DrawCards(deck, 16)
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
		d, err := NewDealer(seed)
		assert.Nil(t, err)
		remaining, drawn, err := d.DrawCards(deck, 16)
		assert.Equal(t, 16, len(remaining))
		assert.Equal(t, 16, len(drawn))
		for i, drawnCard1 := range drawn {
			for _, drawnCard2 := range drawn[i+1:] {
				assert.NotEqual(t, drawnCard1, drawnCard2)
			}
		}
	})
}

func TestDeterministicDealerDeterministic(t *testing.T) {
	seed := []byte{23, 42}

	t.Run("deal a full hand", func(t *testing.T) {
		deck := NewCardDeck()
		d, err := NewDealer(seed)
		assert.Nil(t, err)
		var drawn []Card

		deck, drawn, err = d.DrawCards(deck, 3)
		assert.Nil(t, err)
		assert.Equal(t, 29, len(deck))
		assert.Equal(t, Card{Card10, SuitAcorns}, drawn[0])
		assert.Equal(t, Card{CardAce, SuitBells}, drawn[1])
		assert.Equal(t, Card{CardQueen, SuitHearts}, drawn[2])

		deck, drawn, err = d.DrawCards(deck, 3)
		assert.Nil(t, err)
		assert.Equal(t, 26, len(deck))
		assert.Equal(t, Card{Card10, SuitBells}, drawn[0])
		assert.Equal(t, Card{Card8, SuitBells}, drawn[1])
		assert.Equal(t, Card{Card10, SuitHearts}, drawn[2])

		deck, drawn, err = d.DrawCards(deck, 3)
		assert.Nil(t, err)
		assert.Equal(t, 23, len(deck))
		assert.Equal(t, Card{CardQueen, SuitAcorns}, drawn[0])
		assert.Equal(t, Card{CardJack, SuitLeaves}, drawn[1])
		assert.Equal(t, Card{Card9, SuitAcorns}, drawn[2])

		deck, drawn, err = d.DrawCards(deck, 2)
		assert.Nil(t, err)
		assert.Equal(t, 21, len(deck))
		assert.Equal(t, Card{CardAce, SuitHearts}, drawn[0])
		assert.Equal(t, Card{Card7, SuitAcorns}, drawn[1])

		deck, drawn, err = d.DrawCards(deck, 4)
		assert.Nil(t, err)
		assert.Equal(t, 17, len(deck))
		assert.Equal(t, Card{CardJack, SuitBells}, drawn[0])
		assert.Equal(t, Card{CardKing, SuitLeaves}, drawn[1])
		assert.Equal(t, Card{Card7, SuitHearts}, drawn[2])
		assert.Equal(t, Card{Card9, SuitBells}, drawn[3])

		deck, drawn, err = d.DrawCards(deck, 4)
		assert.Nil(t, err)
		assert.Equal(t, 13, len(deck))
		assert.Equal(t, Card{CardQueen, SuitBells}, drawn[0])
		assert.Equal(t, Card{Card8, SuitLeaves}, drawn[1])
		assert.Equal(t, Card{CardQueen, SuitLeaves}, drawn[2])
		assert.Equal(t, Card{Card8, SuitHearts}, drawn[3])

		deck, drawn, err = d.DrawCards(deck, 4)
		assert.Nil(t, err)
		assert.Equal(t, 9, len(deck))
		assert.Equal(t, Card{Card9, SuitHearts}, drawn[0])
		assert.Equal(t, Card{Card7, SuitLeaves}, drawn[1])
		assert.Equal(t, Card{CardKing, SuitAcorns}, drawn[2])
		assert.Equal(t, Card{Card8, SuitAcorns}, drawn[3])
	})
}
