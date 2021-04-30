package replay

import (
	"encoding/json"

	"github.com/horazont/webskat/internal/skat"
)

type ActionDeclare struct {
	GameType          skat.GameType
	AnnounceModifiers skat.GameModifier
	CardsToPush       skat.CardSet
}

func (a *ActionDeclare) Apply(g *skat.GameState, player int) error {
	return g.Declare(player, a.GameType, a.AnnounceModifiers, a.CardsToPush)
}

func (a *ActionDeclare) Kind() ActionKind {
	return ActionKindDeclare
}

func DecodeActionDeclare(msg []byte) (result *ActionDeclare, err error) {
	result = &ActionDeclare{}
	err = json.Unmarshal(msg, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
