package skat

import (
	"errors"
)

const (
	// Player has not taken part in bidding (yet)
	BidNone = -1

	// Player has passed at 18
	BidPass = 0
)

var (
	ErrBidTooLow = errors.New("bid value too low")
)

type BiddingPlayerState struct {
	LastBid      int
	HasPassedBid bool
}

func NewBiddingPlayerState() BiddingPlayerState {
	return BiddingPlayerState{
		LastBid:      BidNone,
		HasPassedBid: false,
	}
}

type BiddingState struct {
	players          [3]BiddingPlayerState
	declarer         int
	awaitingResponse bool
}

func NewBiddingState() *BiddingState {
	return &BiddingState{
		players: [3]BiddingPlayerState{
			NewBiddingPlayerState(),
			NewBiddingPlayerState(),
			NewBiddingPlayerState(),
		},
		declarer:         PlayerNone,
		awaitingResponse: false,
	}
}

// Returns caller and responder
func (b *BiddingState) evalState() (caller int, responder int) {
	if b.declarer != PlayerNone {
		return PlayerNone, PlayerNone
	}

	forehandPassed := b.players[PlayerInitialForehand].HasPassedBid
	middlehandPassed := b.players[PlayerInitialMiddlehand].HasPassedBid
	rearhandPassed := b.players[PlayerInitialRearhand].HasPassedBid

	if forehandPassed {
		if rearhandPassed {
			if middlehandPassed {
				caller = PlayerNone
			} else {
				caller = PlayerInitialMiddlehand
			}
			responder = PlayerNone
		} else {
			caller = PlayerInitialRearhand
			if middlehandPassed {
				responder = PlayerNone
			} else {
				responder = PlayerInitialMiddlehand
			}
		}
	} else {
		if middlehandPassed {
			if rearhandPassed {
				caller = PlayerInitialForehand
				responder = PlayerNone
			} else {
				caller = PlayerInitialRearhand
				responder = PlayerInitialForehand
			}
		} else {
			caller = PlayerInitialMiddlehand
			responder = PlayerInitialForehand
		}
	}

	return caller, responder
}

func (b *BiddingState) autoconclude() {
	if b.declarer != PlayerNone {
		return
	}
	caller, responder := b.evalState()
	if caller == PlayerNone {
		return
	}
	if responder != PlayerNone {
		return
	}
	if b.players[caller].LastBid != BidNone {
		b.declarer = caller
	}
}

// Return the current caller
//
// Returns PlayerNone iff Done() is true
func (b *BiddingState) Caller() int {
	caller, _ := b.evalState()
	return caller
}

// Return the current responder
//
// Returns PlayerNone iff Done() is true
func (b *BiddingState) Responder() int {
	_, responder := b.evalState()
	return responder
}

func (b *BiddingState) AwaitingResponse() bool {
	return b.awaitingResponse
}

// Return true if the bidding is over
func (b *BiddingState) Done() bool {
	caller, _ := b.evalState()
	return caller == PlayerNone
}

// Return the called game value.
//
// If Done() is false, this returns 0.
func (b *BiddingState) CalledGameValue() int {
	if b.declarer == PlayerNone {
		return 0
	}
	return b.players[b.declarer].LastBid
}

// Returns the declarer
//
// If Done() is false or if all players have passed without placing any bid,
// this returns PlayerNone.
func (b *BiddingState) Declarer() int {
	return b.declarer
}

// Place a bid as the given player.
func (b *BiddingState) Call(player int, value int) error {
	if b.awaitingResponse {
		return ErrNotYourTurn
	}
	caller, _ := b.evalState()
	if caller != player {
		return ErrNotYourTurn
	}
	if value == BidPass {
		b.players[player].HasPassedBid = true
	} else {
		if b.players[player].LastBid >= value {
			return ErrBidTooLow
		}
		b.players[player].LastBid = value
		b.awaitingResponse = true
	}
	b.autoconclude()
	return nil
}

func (b *BiddingState) Respond(player int, hold bool) error {
	if !b.awaitingResponse {
		return ErrNotYourTurn
	}
	caller, responder := b.evalState()
	if responder != player {
		return ErrNotYourTurn
	}
	b.players[player].HasPassedBid = !hold
	if hold {
		b.players[player].LastBid = b.players[caller].LastBid
	}
	b.awaitingResponse = false
	b.autoconclude()
	return nil
}
