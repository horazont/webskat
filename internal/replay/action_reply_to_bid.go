package replay

import (
	"encoding/json"

	"github.com/horazont/webskat/internal/skat"
)

type ActionReplyToBid struct {
	Hold bool `json:"hold"`
}

func (a *ActionReplyToBid) Apply(g *skat.GameState, player int) error {
	return g.RespondToBid(player, a.Hold)
}

func (a *ActionReplyToBid) Kind() ActionKind {
	return ActionKindReplyToBid
}

func DecodeActionReplyToBid(msg []byte) (result *ActionReplyToBid, err error) {
	result = &ActionReplyToBid{}
	err = json.Unmarshal(msg, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
