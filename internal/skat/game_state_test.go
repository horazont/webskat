package skat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testGetInitPhaseGame(t *testing.T) *GameState {
	g := NewGame(false, LeagueScoreDefinition())
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

func testGetPlayingPhaseGame(t *testing.T, gameType GameType) *GameState {
	g := testGetDeclarationPhaseGame(t)
	assert.Nil(t, g.Declare(PlayerInitialMiddlehand, gameType, NoGameModifiers, nil))
	return g
}

func testGetDonePlayingPhaseGame(t *testing.T, gameType GameType) *GameState {
	g := testGetPlayingPhaseGame(t, gameType)
	play := g.Playing()
	for i := 0; i < 10; i = i + 1 {
		// 10 tricks
		player := play.GetCurrentPlayer()
		assert.Nil(t, play.Play(player, play.GetHand(player)[0]))
		for j := 0; j < 2; j = j + 1 {
			// 2 non-forehand players
			player := play.GetCurrentPlayer()
			hand := play.GetHand(player)
			success := false
			for _, card := range hand {
				if play.Play(player, card) == nil {
					success = true
					break
				}
			}
			assert.True(t, success)
		}
	}
	assert.Equal(t, 0, len(g.GetHand(PlayerInitialForehand)))
	assert.Equal(t, 0, len(g.GetHand(PlayerInitialMiddlehand)))
	assert.Equal(t, 0, len(g.GetHand(PlayerInitialRearhand)))
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

	t.Run("reject declare", func(t *testing.T) {
		g := testGetBiddingPhaseGame(t)

		assert.Equal(t, ErrWrongPhase, g.Declare(PlayerInitialMiddlehand, GameTypeBells, NoGameModifiers, []Card{}))
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

		assert.True(t, g.Modifiers().Test(GameModifierHand))
	})

	t.Run("taking skat transfers to hand of declarer", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)
		skat := g.GetSkat()

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		hand := g.GetHand(PlayerInitialMiddlehand)
		assert.Equal(t, 12, len(hand))
		assert.Equal(t, skat[0], hand[10])
		assert.Equal(t, skat[1], hand[11])
	})

	t.Run("taking skat clears Hand modifier", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		assert.False(t, g.Modifiers().Test(GameModifierHand))
	})

	t.Run("reject declare without push while skat is on hand", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		assert.Equal(t, ErrInvalidPush, g.Declare(PlayerInitialMiddlehand, GameTypeHearts, NoGameModifiers, []Card{}))
	})

	t.Run("reject declare from non-declarer", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		assert.Equal(t, ErrNotYourTurn, g.Declare(PlayerInitialForehand, GameTypeHearts, NoGameModifiers, []Card{}))
	})

	t.Run("reject declare with non-announcable states", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)
		skat := g.GetSkat()

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		assert.Equal(t, ErrInvalidGame, g.Declare(PlayerInitialMiddlehand, GameTypeHearts, GameModifierSchwarz, []Card{skat[0], skat[1]}))
	})

	t.Run("reject declare with invalid states for game type", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)
		skat := g.GetSkat()

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		assert.Equal(t, ErrInvalidGame, g.Declare(PlayerInitialMiddlehand, GameTypeNull, GameModifierSchwarzAnnounced.Normalized(), []Card{skat[0], skat[1]}))
	})

	t.Run("declare with valid game type transitions to playing phase", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)

		assert.Nil(t, g.Declare(PlayerInitialMiddlehand, GameTypeHearts, NoGameModifiers, nil))
		assert.Equal(t, PhasePlaying, g.Phase())
	})

	t.Run("declare with valid game type and push processes push", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)
		skat := g.GetSkat()
		handBefore := g.GetHand(PlayerInitialMiddlehand)

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))
		assert.Nil(t, g.Declare(PlayerInitialMiddlehand, GameTypeHearts, NoGameModifiers, skat))
		handAfter := g.GetHand(PlayerInitialMiddlehand)
		assert.Equal(t, 10, len(handAfter))
		assert.Equal(t, handBefore, handAfter)
	})

	t.Run("reject declare with push with cards not in hand", func(t *testing.T) {
		g := testGetDeclarationPhaseGame(t)
		handBefore := g.GetHand(PlayerInitialMiddlehand)
		otherHand := g.GetHand(PlayerInitialForehand)

		assert.Nil(t, g.TakeSkat(PlayerInitialMiddlehand))

		handWithSkat := g.GetHand(PlayerInitialMiddlehand)
		assert.Equal(t, ErrInvalidPush, g.Declare(PlayerInitialMiddlehand, GameTypeHearts, NoGameModifiers, []Card{handBefore[0], otherHand[0]}))
		assert.Equal(t, PhaseDeclaration, g.Phase())
		handAfter := g.GetHand(PlayerInitialMiddlehand)
		assert.Equal(t, 12, len(handAfter))
		assert.Equal(t, handWithSkat, handAfter)
	})
}

func TestGameStatePlayingPhase(t *testing.T) {
	t.Run("playing tricks mutates frontend hand", func(t *testing.T) {
		g := testGetPlayingPhaseGame(t, GameTypeHearts)
		handBefore := g.GetHand(PlayerInitialForehand)
		assert.Nil(t, g.Playing().Play(PlayerInitialForehand, handBefore[0]))
		handAfter := g.GetHand(PlayerInitialForehand)
		assert.Equal(t, 9, len(handAfter))
		assert.False(t, handAfter.Contains(handBefore[0]))
	})

	t.Run("playing state is initialized correctly", func(t *testing.T) {
		g := testGetPlayingPhaseGame(t, GameTypeHearts)
		for i := PlayerInitialForehand; i <= PlayerInitialRearhand; i = i + 1 {
			assert.Equal(t, g.GetHand(i), g.Playing().GetHand(i))
		}
		assert.Equal(t, PlayerInitialMiddlehand, g.Playing().Declarer())
		assert.Equal(t, 2, len(g.Playing().GetWonCards(g.Playing().Declarer())))
		assert.Equal(t, PlayerInitialForehand, g.Playing().GetCurrentPlayer())
		assert.Equal(t, GameTypeHearts, g.Playing().GameType())
	})
}

func TestGameStatePlayingPhaseCompleted(t *testing.T) {
	t.Run("evaluate game: standard bells", func(t *testing.T) {
		g := testGetDonePlayingPhaseGame(t, GameTypeBells)
		assert.Nil(t, g.EvaluateGame())
		assert.Equal(t, PhaseScored, g.Phase())
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialForehand)))
		assert.Equal(t, 12, len(g.GetHand(PlayerInitialMiddlehand)))
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialRearhand)))
		assert.Equal(t, 40, g.GetScore(PlayerInitialForehand))
		assert.Equal(t, -54, g.GetScore(PlayerInitialMiddlehand))
		assert.Equal(t, 40, g.GetScore(PlayerInitialRearhand))
		assert.Equal(t, LossReasonNotEnoughPoints, g.GetLossReason())
	})

	t.Run("evaluate game: standard hearts", func(t *testing.T) {
		g := testGetDonePlayingPhaseGame(t, GameTypeHearts)
		assert.Nil(t, g.EvaluateGame())
		assert.Equal(t, PhaseScored, g.Phase())
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialForehand)))
		assert.Equal(t, 12, len(g.GetHand(PlayerInitialMiddlehand)))
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialRearhand)))
		assert.Equal(t, 40, g.GetScore(PlayerInitialForehand))
		assert.Equal(t, -60, g.GetScore(PlayerInitialMiddlehand))
		assert.Equal(t, 40, g.GetScore(PlayerInitialRearhand))
		assert.Equal(t, LossReasonNotEnoughPoints, g.GetLossReason())
	})

	t.Run("evaluate game: standard leaves", func(t *testing.T) {
		g := testGetDonePlayingPhaseGame(t, GameTypeLeaves)
		assert.Nil(t, g.EvaluateGame())
		assert.Equal(t, PhaseScored, g.Phase())
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialForehand)))
		assert.Equal(t, 12, len(g.GetHand(PlayerInitialMiddlehand)))
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialRearhand)))
		assert.Equal(t, 40, g.GetScore(PlayerInitialForehand))
		assert.Equal(t, -88, g.GetScore(PlayerInitialMiddlehand))
		assert.Equal(t, 40, g.GetScore(PlayerInitialRearhand))
		assert.Equal(t, LossReasonNotEnoughPoints, g.GetLossReason())
	})

	t.Run("evaluate game: standard acorns", func(t *testing.T) {
		g := testGetDonePlayingPhaseGame(t, GameTypeAcorns)
		assert.Nil(t, g.EvaluateGame())
		assert.Equal(t, PhaseScored, g.Phase())
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialForehand)))
		assert.Equal(t, 12, len(g.GetHand(PlayerInitialMiddlehand)))
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialRearhand)))
		assert.Equal(t, 0, g.GetScore(PlayerInitialForehand))
		assert.Equal(t, -72, g.GetScore(PlayerInitialMiddlehand))
		assert.Equal(t, 0, g.GetScore(PlayerInitialRearhand))
		assert.Equal(t, "", g.GetLossReason())
	})

	t.Run("evaluate game: standard grand", func(t *testing.T) {
		g := testGetDonePlayingPhaseGame(t, GameTypeGrand)
		assert.Nil(t, g.EvaluateGame())
		assert.Equal(t, PhaseScored, g.Phase())
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialForehand)))
		assert.Equal(t, 12, len(g.GetHand(PlayerInitialMiddlehand)))
		assert.Equal(t, 10, len(g.GetHand(PlayerInitialRearhand)))
		assert.Equal(t, 40, g.GetScore(PlayerInitialForehand))
		assert.Equal(t, -192, g.GetScore(PlayerInitialMiddlehand))
		assert.Equal(t, 40, g.GetScore(PlayerInitialRearhand))
		assert.Equal(t, LossReasonNotEnoughPoints, g.GetLossReason())
	})
}
