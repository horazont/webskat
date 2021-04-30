package replay

import (
	"encoding/json"

	"github.com/horazont/webskat/internal/skat"
)

type ActionSetSeed struct {
	Seed skat.Seed `json:"seed"`
}

func (a *ActionSetSeed) Apply(g *skat.GameState, player int) error {
	if player == skat.PlayerNone {
		// dealer
		return g.SetDealerSeed(a.Seed)
	} else {
		return g.SetSeed(player, a.Seed)
	}
}

func (a *ActionSetSeed) Kind() ActionKind {
	return ActionKindSetSeed
}

func DecodeActionSetSeed(msg []byte) (result *ActionSetSeed, err error) {
	result = &ActionSetSeed{}
	err = json.Unmarshal(msg, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func SetSeed(seed []byte) *ActionSetSeed {
	return &ActionSetSeed{Seed: seed}
}
