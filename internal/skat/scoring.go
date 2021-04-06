package skat

import (
	"sort"
)

const (
	LossReasonNotEnoughPoints = "not_enough_points"
	LossReasonNoSchneider     = "no_schneider"
	LossReasonNoSchwarz       = "no_schwarz"
	LossReasonOverbid         = "overbid"
)

type ScoreFormula struct {
	Offset int
	Factor int
}

func (f ScoreFormula) Apply(value int) int {
	return f.Offset + f.Factor*value
}

type ScoreDefinition struct {
	DeclarerWin  ScoreFormula
	DeclarerLoss ScoreFormula
	DefenderWin  ScoreFormula
	DefenderLoss ScoreFormula
}

func StandardScoreDefinition() *ScoreDefinition {
	return &ScoreDefinition{
		DeclarerWin: ScoreFormula{
			Offset: 0,
			Factor: 1,
		},
		DeclarerLoss: ScoreFormula{
			Offset: 0,
			Factor: -2,
		},
		DefenderWin: ScoreFormula{
			Offset: 0,
			Factor: 0,
		},
		DefenderLoss: ScoreFormula{
			Offset: 0,
			Factor: 0,
		},
	}
}

func LeagueScoreDefinition() *ScoreDefinition {
	return &ScoreDefinition{
		DeclarerWin: ScoreFormula{
			Offset: 50,
			Factor: 1,
		},
		DeclarerLoss: ScoreFormula{
			Offset: 0,
			Factor: -2,
		},
		DefenderWin: ScoreFormula{
			Offset: 40,
			Factor: 0,
		},
		DefenderLoss: ScoreFormula{
			Offset: 0,
			Factor: 0,
		},
	}
}

func (sd *ScoreDefinition) CalculateScore(gameValue int, declarer int, declarerWon bool) [3]int {
	result := [3]int{0, 0, 0}
	defender1 := (declarer + 1) % 3
	defender2 := (defender1 + 1) % 3

	defenderFormula := sd.DefenderWin
	declarerFormula := sd.DeclarerLoss
	if declarerWon {
		defenderFormula = sd.DefenderLoss
		declarerFormula = sd.DeclarerLoss
	}

	result[declarer] = declarerFormula.Apply(gameValue)
	result[defender1] = defenderFormula.Apply(gameValue)
	result[defender2] = result[defender1]

	return result
}

func effectivePower(c Card, gameType GameType) int {
	suit := c.EffectiveSuit(gameType)
	if suit == EffectiveSuitTrumps {
		return c.RelativePower(gameType)
	}
	return -1
}

var (
	referenceOrder = CardSet{
		CardJack.As(SuitAcorns),
		CardJack.As(SuitLeaves),
		CardJack.As(SuitHearts),
		CardJack.As(SuitBells),
		CardAce.As(SuitBells),
		Card10.As(SuitBells),
		CardKing.As(SuitBells),
		CardQueen.As(SuitBells),
		Card9.As(SuitBells),
		Card8.As(SuitBells),
		Card7.As(SuitBells),
	}
)

const (
	referenceGameType = GameTypeBells
)

func (cs CardSet) GetMatadorsJackStrength(gameType GameType) int {
	if gameType == GameTypeNull {
		return 0
	}

	max := 11
	refOrder := referenceOrder
	refGameType := referenceGameType
	if gameType == GameTypeGrand {
		max = 4
		refOrder = refOrder[:4]
		refGameType = GameTypeGrand
	}

	tmp := cs.Copy()
	sort.SliceStable(
		tmp[:],
		func(i, j int) bool {
			// must return true if tmp[i] < tmp[j]
			// we want to have reverse ordering (highest first), so we invert
			// the scores before the final comparison
			iPower := -effectivePower(tmp[i], gameType)
			jPower := -effectivePower(tmp[j], gameType)
			return iPower < jPower
		},
	)
	if tmp[0].EffectiveSuit(gameType) != EffectiveSuitTrumps {
		return max
	}
	firstTrumpPower := tmp[0].RelativePower(gameType)
	if firstTrumpPower == refOrder[0].RelativePower(refGameType) {
		// game "with"
		for i, card := range tmp {
			if card.EffectiveSuit(gameType) != EffectiveSuitTrumps {
				return i
			}
			if card.RelativePower(gameType) != refOrder[i].RelativePower(refGameType) {
				return i
			}
		}
	} else {
		// game "without"
		for i, refCard := range refOrder {
			if firstTrumpPower == refCard.RelativePower(refGameType) {
				return i
			}
		}
	}
	return max
}

func CalculateGameValue(initialDeclarerHand CardSet, gameType GameType, modifiers GameModifier) (base int, factor int) {
	factor = 1
	switch gameType {
	case GameTypeNull:
		isHand := modifiers.Test(GameModifierHand)
		isOuvert := modifiers.Test(GameModifierOuvert)
		if isHand {
			if isOuvert {
				return 59, factor
			} else {
				return 35, factor
			}
		} else {
			if isOuvert {
				return 46, factor
			} else {
				return 23, factor
			}
		}
	case GameTypeBells, GameTypeHearts, GameTypeLeaves, GameTypeAcorns, GameTypeGrand:
		switch gameType {
		case GameTypeBells:
			base = 9
		case GameTypeHearts:
			base = 10
		case GameTypeLeaves:
			base = 11
		case GameTypeAcorns:
			base = 12
		case GameTypeGrand:
			base = 24
		}
		factor = 1 + initialDeclarerHand.GetMatadorsJackStrength(gameType)
		if modifiers.Test(GameModifierHand) {
			factor = factor + 1
			if modifiers.Test(GameModifierSchneiderAnnounced) {
				factor = factor + 1
			}
			if modifiers.Test(GameModifierSchwarzAnnounced) {
				factor = factor + 1
			}
		}
		if modifiers.Test(GameModifierSchneider) {
			factor = factor + 1
		}
		if modifiers.Test(GameModifierSchwarz) {
			factor = factor + 1
		}
		if modifiers.Test(GameModifierOuvert) {
			factor = factor + 1
		}
		return base, factor
	}
	return base, factor
}

func EvaluateWonCards(wonCards [3]CardSet, declarer int) (modifiers GameModifier, declarerScore int, defenderScore int) {
	declarerScore = wonCards[declarer].Value()
	defender1 := (declarer + 1) % 3
	defender2 := (defender1 + 1) % 3
	defenderScore = wonCards[defender1].Value() + wonCards[defender2].Value()

	if defenderScore <= 30 || declarerScore <= 30 {
		modifiers = modifiers.With(GameModifierSchneider)
	}

	if len(wonCards[declarer]) <= 2 || len(wonCards[defender1])+len(wonCards[defender2]) <= 2 {
		modifiers = modifiers.With(GameModifierSchwarz)
	}

	return modifiers, declarerScore, defenderScore
}

func EvaluateGame(gameBaseValue int, gameValueFactor int, declarerScore int, declarerBid int) (declarerWon bool, gameValue int, lossReason string) {
	gameValue = gameBaseValue * gameValueFactor
	if gameValue > declarerBid {
		// overbid!
		gameValue = gameBaseValue * ((declarerBid + gameBaseValue - 1) / gameBaseValue)
		return false, gameValue, LossReasonOverbid
	}

	if declarerScore <= 60 {
		return false, gameValue, LossReasonNotEnoughPoints
	}
	return true, gameValue, ""
}
