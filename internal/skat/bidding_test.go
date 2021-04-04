package skat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBiddingOrder(t *testing.T) {
	t.Run("initial caller and responder", func(t *testing.T) {
		b := NewBiddingState()
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())
	})

	t.Run("initial state", func(t *testing.T) {
		b := NewBiddingState()
		assert.Equal(t, 0, b.CalledGameValue())
		assert.False(t, b.Done())
	})

	t.Run("reject initial bid from non-caller", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialForehand, 18)
		assert.Equal(t, ErrNotYourTurn, err)

		err = b.Call(PlayerInitialRearhand, 18)
		assert.Equal(t, ErrNotYourTurn, err)
	})

	t.Run("reject initial response", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Respond(PlayerInitialForehand, true)
		assert.Equal(t, ErrNotYourTurn, err)

		err = b.Respond(PlayerInitialMiddlehand, true)
		assert.Equal(t, ErrNotYourTurn, err)

		err = b.Respond(PlayerInitialRearhand, true)
		assert.Equal(t, ErrNotYourTurn, err)
	})

	t.Run("reject response after initial call from non-responder", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialMiddlehand, 18)
		assert.Nil(t, err)

		err = b.Respond(PlayerInitialMiddlehand, true)
		assert.Equal(t, ErrNotYourTurn, err)

		err = b.Respond(PlayerInitialRearhand, true)
		assert.Equal(t, ErrNotYourTurn, err)
	})

	t.Run("reject call after initial call", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialMiddlehand, 18)
		assert.Nil(t, err)

		err = b.Call(PlayerInitialForehand, BidPass)
		assert.Equal(t, ErrNotYourTurn, err)

		err = b.Call(PlayerInitialMiddlehand, BidPass)
		assert.Equal(t, ErrNotYourTurn, err)

		err = b.Call(PlayerInitialRearhand, BidPass)
		assert.Equal(t, ErrNotYourTurn, err)
	})

	t.Run("reject lower or equal bid", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialMiddlehand, 18)
		assert.Nil(t, err)

		err = b.Respond(PlayerInitialForehand, true)
		assert.Nil(t, err)

		err = b.Call(PlayerInitialMiddlehand, 18)
		assert.Equal(t, ErrBidTooLow, err)
		assert.False(t, b.AwaitingResponse())
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())

		err = b.Call(PlayerInitialMiddlehand, 17)
		assert.Equal(t, ErrBidTooLow, err)
		assert.False(t, b.AwaitingResponse())
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())
	})

	t.Run("18-pass-pass: middlehand takes with 18", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialMiddlehand, 18)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.True(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Respond(PlayerInitialForehand, false)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialMiddlehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialRearhand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerNone, b.Caller())
		assert.Equal(t, PlayerNone, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.True(t, b.Done())
		assert.Equal(t, PlayerInitialMiddlehand, b.Declarer())
		assert.Equal(t, 18, b.CalledGameValue())
	})

	t.Run("18-hold-pass-pass: forehand takes with 18", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialMiddlehand, 18)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.True(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Respond(PlayerInitialForehand, true)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialMiddlehand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialRearhand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerNone, b.Caller())
		assert.Equal(t, PlayerNone, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.True(t, b.Done())
		assert.Equal(t, PlayerInitialForehand, b.Declarer())
		assert.Equal(t, 18, b.CalledGameValue())
	})

	t.Run("18-hold-20-hold-pass-pass: forehand takes with 20", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialMiddlehand, 18)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.True(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Respond(PlayerInitialForehand, true)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialMiddlehand, 20)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.True(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Respond(PlayerInitialForehand, true)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialMiddlehand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialRearhand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerNone, b.Caller())
		assert.Equal(t, PlayerNone, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.True(t, b.Done())
		assert.Equal(t, PlayerInitialForehand, b.Declarer())
		assert.Equal(t, 20, b.CalledGameValue())
	})

	t.Run("18-hold-pass-20-pass: rearhand takes with 20", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialMiddlehand, 18)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.True(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Respond(PlayerInitialForehand, true)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialMiddlehand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialMiddlehand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialRearhand, 20)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.True(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Respond(PlayerInitialForehand, false)
		assert.Nil(t, err)
		assert.Equal(t, PlayerNone, b.Caller())
		assert.Equal(t, PlayerNone, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.True(t, b.Done())
		assert.Equal(t, PlayerInitialRearhand, b.Declarer())
		assert.Equal(t, 20, b.CalledGameValue())
	})

	t.Run("pass-18-pass: rearhand takes with 18", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialMiddlehand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialRearhand, 18)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.True(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Respond(PlayerInitialForehand, false)
		assert.Nil(t, err)
		assert.Equal(t, PlayerNone, b.Caller())
		assert.Equal(t, PlayerNone, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.True(t, b.Done())
		assert.Equal(t, PlayerInitialRearhand, b.Declarer())
		assert.Equal(t, 18, b.CalledGameValue())
	})

	t.Run("pass-18-hold-pass: forehand takes with 18 after initial pass", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialMiddlehand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialRearhand, 18)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.True(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Respond(PlayerInitialForehand, true)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialRearhand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerNone, b.Caller())
		assert.Equal(t, PlayerNone, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.True(t, b.Done())
		assert.Equal(t, PlayerInitialForehand, b.Declarer())
		assert.Equal(t, 18, b.CalledGameValue())
	})

	t.Run("pass-18-hold-20-pass: rearhand takes with 20 after initial pass", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialMiddlehand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialRearhand, 18)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.True(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Respond(PlayerInitialForehand, true)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialRearhand, 20)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.True(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Respond(PlayerInitialForehand, false)
		assert.Nil(t, err)
		assert.Equal(t, PlayerNone, b.Caller())
		assert.Equal(t, PlayerNone, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.True(t, b.Done())
		assert.Equal(t, PlayerInitialRearhand, b.Declarer())
		assert.Equal(t, 20, b.CalledGameValue())
	})

	t.Run("pass-pass-pass: nobody takes", func(t *testing.T) {
		b := NewBiddingState()
		var err error

		err = b.Call(PlayerInitialMiddlehand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialRearhand, b.Caller())
		assert.Equal(t, PlayerInitialForehand, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialRearhand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerInitialForehand, b.Caller())
		assert.Equal(t, PlayerNone, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.False(t, b.Done())

		err = b.Call(PlayerInitialForehand, BidPass)
		assert.Nil(t, err)
		assert.Equal(t, PlayerNone, b.Caller())
		assert.Equal(t, PlayerNone, b.Responder())
		assert.False(t, b.AwaitingResponse())
		assert.True(t, b.Done())
		assert.Equal(t, PlayerNone, b.Declarer())
		assert.Equal(t, 0, b.CalledGameValue())
	})
}
