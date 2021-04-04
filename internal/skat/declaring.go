package skat

type DeclarationState struct {
	declarer        int
	calledGameValue int
	stateModifiers  GameModifier
	skat            []Card
}
