package replay

import (
	"encoding/json"
	"errors"

	"github.com/horazont/webskat/internal/skat"
)

type ActionKind string

const (
	// Init phase
	ActionKindSetSeed ActionKind = "set_seed"

	// Bidding phase
	ActionKindCallBid    ActionKind = "bid_call"
	ActionKindReplyToBid ActionKind = "bid_reply"

	// Declaration phase
	ActionKindTakeSkat ActionKind = "take_skat"
	ActionKindDeclare  ActionKind = "declare"

	// Playing phase
	ActionKindPlayCard ActionKind = "play"

	// Declaration / Playing phases
	ActionKindResign ActionKind = "resign"

	// Bidding / Declaration / Playing phases
	ActionKindPeek ActionKind = "peek"
)

const (
	ActionCallBidParameterPass    = "pass"
	ActionReplyToBidParameterPass = "pass"
	ActionReplyToBidParameterHold = "hold"
)

var (
	ErrUnknownAction = errors.New("unknown action")
)

type Action interface {
	Apply(g *skat.GameState, player int) error
	Kind() ActionKind
}

type intermediateAction struct {
	Kind          string          `json:"kind"`
	ActionPayload json.RawMessage `json:"spec"`
}

func ActionFromJSON(dec *json.Decoder) (Action, error) {
	ia := &intermediateAction{}
	err := dec.Decode(ia)
	if err != nil {
		return nil, err
	}

	switch ActionKind(ia.Kind) {
	case ActionKindSetSeed:
		return DecodeActionSetSeed(ia.ActionPayload)
	case ActionKindCallBid:
		return DecodeActionCallBid(ia.ActionPayload)
	case ActionKindReplyToBid:
		return DecodeActionReplyToBid(ia.ActionPayload)
	case ActionKindTakeSkat:
		return DecodeActionTakeSkat(ia.ActionPayload)
	case ActionKindDeclare:
		return DecodeActionDeclare(ia.ActionPayload)
	case ActionKindPlayCard:
		return DecodeActionPlayCard(ia.ActionPayload)
	}

	return nil, nil
}

func ActionToJSON(action Action, enc *json.Encoder) error {
	data, err := json.Marshal(action)
	if err != nil {
		return err
	}

	ia := intermediateAction{
		Kind:          string(action.Kind()),
		ActionPayload: data,
	}
	return enc.Encode(&ia)
}
