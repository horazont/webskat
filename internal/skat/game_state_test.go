package skat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testGetInitPhaseGame(t *testing.T) *GameState {
	g := NewGame(false)
	assert.Equal(t, PhaseInit, g.Phase())
	return g
}

func testGetBiddingPhaseGame(t *testing.T) *GameState {
	g := testGetInitPhaseGame(t)
	assert.Nil(t, g.SetSeed(PlayerInitialForehand, []byte{}))
	assert.Nil(t, g.SetSeed(PlayerInitialMiddlehand, []byte{}))
	assert.Nil(t, g.SetSeed(PlayerInitialRearhand, []byte{}))
	assert.Nil(t, g.SetServerSeed([]byte{}))
	assert.Nil(t, g.Deal())
	assert.Equal(t, PhaseBidding, g.Phase())
	return g
}

func testGetDeclarationPhaseGame(t *testing.T) *GameState {
	g := testGetBiddingPhaseGame(t)
	assert.Nil(t, g.Bidding().Call(PlayerInitialMiddlehand, 18))
	assert.Nil(t, g.Bidding().Respond(PlayerInitialForehand, false))
	assert.Nil(t, g.Bidding().Call(PlayerInitialRearhand, BidPass))
	assert.True(t, g.Bidding().Done())
	assert.Nil(t, g.ConcludeBidding())
	assert.Equal(t, PhaseDeclaration, g.Phase())
	return g
}

func TestGameStateInitialPhase(t *testing.T) {
	t.Run("rejects dealing without all seeds", func(t *testing.T) {
		g := testGetInitPhaseGame(t)

		var err error

		err = g.Deal()
		assert.Equal(t, ErrMissingSeed, err)

		g.SetSeed(PlayerInitialForehand, []byte{})

		err = g.Deal()
		assert.Equal(t, ErrMissingSeed, err)

		g.SetSeed(PlayerInitialMiddlehand, []byte{})

		err = g.Deal()
		assert.Equal(t, ErrMissingSeed, err)

		g.SetSeed(PlayerInitialRearhand, []byte{})

		err = g.Deal()
		assert.Equal(t, ErrMissingSeed, err)

		g.SetServerSeed([]byte{})

		err = g.Deal()
		assert.Nil(t, err)
	})

	t.Run("reject transition to declaration phase", func(t *testing.T) {
		g := testGetInitPhaseGame(t)
		assert.Equal(t, ErrWrongPhase, g.ConcludeBidding())
	})
}

func TestGameStateBiddingPhase(t *testing.T) {
	t.Run("hands are dealt", func(t *testing.T) {
		g := testGetBiddingPhaseGame(t)

		assert.Equal(t, 10, len(g.GetHand(PlayerInitialForehand)))
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialMiddlehand)))
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialRearhand)))
		assert.Equal(t, 2, len(g.GetSkat()))
	})

	t.Run("bidding is initialized", func(t *testing.T) {
		g := testGetBiddingPhaseGame(t)

		assert.NotNil(t, g.Bidding())
	})

	t.Run("reject transition to declaration phase if bidding not done", func(t *testing.T) {
		g := testGetBiddingPhaseGame(t)

		assert.Equal(t, ErrBiddingNotDone, g.ConcludeBidding())
	})

	t.Run("reject dealing a new hand", func(t *testing.T) {
		g := testGetBiddingPhaseGame(t)

		assert.Equal(t, ErrWrongPhase, g.Deal())
	})

	t.Run("reject changing seeds", func(t *testing.T) {
		g := testGetBiddingPhaseGame(t)

		assert.Equal(t, ErrWrongPhase, g.SetSeed(PlayerInitialForehand, []byte{}))
		assert.Equal(t, ErrWrongPhase, g.SetServerSeed([]byte{}))
	})

	t.Run("transition to declaration phase when bidding is concluded", func(t *testing.T) {
		g := testGetBiddingPhaseGame(t)

		assert.Nil(t, g.Bidding().Call(PlayerInitialMiddlehand, 18))
		assert.Nil(t, g.Bidding().Respond(PlayerInitialForehand, false))
		assert.Nil(t, g.Bidding().Call(PlayerInitialRearhand, BidPass))
		assert.True(t, g.Bidding().Done())

		assert.Nil(t, g.ConcludeBidding())
		assert.Equal(t, PhaseDeclaration, g.Phase())
	})
}

func TestGameStateDeclarationPhase(t *testing.T) {
	t.Run("reject taking the skat with a non-declarer", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)

		assert.Equal(t, ErrNotYourTurn, g.TakeSkat(PlayerInitialForehand))
		assert.Equal(t, ErrNotYourTurn, g.TakeSkat(PlayerInitialRearhand))

		assert.Equal(t, 2, len(g.GetSkat()))
	})

	t.Run("initially has Hand modifier", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)

		assert.Equal(t, GameModifierHand, g.GetStateModifiers())
	})

	t.Run("taking skat transfers to hand of declarer", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)
		skat := g.GetSkat()

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		assert.Equal(t, 0, len(g.GetSkat()))
		hand := g.GetHand(PlayerInitialMiddlehand)
		assert.Equal(t, 12, len(hand))
		assert.Equal(t, skat[0], hand[10])
		assert.Equal(t, skat[1], hand[11])
	})

	t.Run("taking skat clears Hand modifier", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		assert.Equal(t, NoGameModifiers, g.GetStateModifiers())
	})

	t.Run("reject push cards if skat not taken", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)

		assert.Equal(t, ErrWrongPhase, g.PushCards(PlayerInitialMiddlehand, Card{}, Card{}))
	})

	t.Run("reject push cards by non-declarer", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)

		assert.Equal(t, ErrNotYourTurn, g.PushCards(PlayerInitialForehand, Card{}, Card{}))
		assert.Equal(t, ErrNotYourTurn, g.PushCards(PlayerInitialRearhand, Card{}, Card{}))
	})

	t.Run("allow pushing the skat", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)
		skat := g.GetSkat()
		handBefore := g.GetHand(PlayerInitialMiddlehand)

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		assert.Nil(t, g.PushCards(PlayerInitialMiddlehand, skat[0], skat[1]))
		handAfter := g.GetHand(PlayerInitialMiddlehand)
		assert.Equal(t, 10, len(handAfter))
		assert.Equal(t, handBefore, handAfter)
	})

	t.Run("allow pushing cards other than the skat", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)
		skat := g.GetSkat()
		handBefore := g.GetHand(PlayerInitialMiddlehand)

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		assert.Nil(t, g.PushCards(PlayerInitialMiddlehand, handBefore[0], handBefore[1]))
		handAfter := g.GetHand(PlayerInitialMiddlehand)
		assert.Equal(t, 10, len(handAfter))
		assert.Equal(t, append(handBefore[2:], skat...), handAfter)
	})

	t.Run("reject declare while skat is on hand", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		assert.Equal(t, ErrWrongPhase, g.Declare(PlayerInitialMiddlehand, GameTypeHearts, NoGameModifiers))
	})
}
