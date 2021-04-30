package replay

import (
	"encoding/json"

	"github.com/horazont/webskat/internal/skat"
)

type ActionPlayCard struct {
	Card skat.Card `json:"card"`
}

func (a *ActionPlayCard) Apply(g *skat.GameState, player int) error {
	return g.PlayCard(player, a.Card)
}

func (a *ActionPlayCard) Kind() ActionKind {
	return ActionKindPlayCard
}

func DecodeActionPlayCard(msg []byte) (result *ActionPlayCard, err error) {
	result = &ActionPlayCard{}
	err = json.Unmarshal(msg, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
