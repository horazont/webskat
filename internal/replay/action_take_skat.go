package replay

import (
	"encoding/json"

	"github.com/horazont/webskat/internal/skat"
)

type ActionTakeSkat struct {
}

func (a *ActionTakeSkat) Apply(g *skat.GameState, player int) error {
	return g.TakeSkat(player)
}

func (a *ActionTakeSkat) Kind() ActionKind {
	return ActionKindTakeSkat
}

func DecodeActionTakeSkat(msg []byte) (result *ActionTakeSkat, err error) {
	result = &ActionTakeSkat{}
	err = json.Unmarshal(msg, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
