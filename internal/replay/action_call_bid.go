package replay

import (
	"encoding/json"

	"github.com/horazont/webskat/internal/skat"
)

type ActionCallBid struct {
	Value int `json:"value"`
}

func (a *ActionCallBid) Apply(g *skat.GameState, player int) error {
	return g.CallBid(player, a.Value)
}

func (a *ActionCallBid) Kind() ActionKind {
	return ActionKindCallBid
}

func DecodeActionCallBid(msg []byte) (result *ActionCallBid, err error) {
	result = &ActionCallBid{}
	err = json.Unmarshal(msg, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
