package safety

import (
	"fmt"

	"github.com/shagston/routerpilot/sdk/types"
)

// Guard определяет интерфейс для проверки безопасности плана
type Guard interface {
	Validate(plan types.Plan) (bool, error)
}

// SimpleSafetyGuard реализует базовую проверку на основе уровня риска
type SimpleSafetyGuard struct {
	// AllowedRisk определяет максимальный уровень риска, который можно выполнять без подтверждения
	AllowedRisk types.RiskLevel
}

func NewSimpleSafetyGuard(allowedRisk types.RiskLevel) *SimpleSafetyGuard {
	return &SimpleSafetyGuard{
		AllowedRisk: allowedRisk,
	}
}

// Validate возвращает true, если план безопасен (или риск допустим), и false, если требуется подтверждение
func (g *SimpleSafetyGuard) Validate(plan types.Plan) (bool, error) {
	if plan.Risk == "" {
		return false, fmt.Errorf("plan risk level is not specified")
	}

	return riskRank(plan.Risk) <= riskRank(g.AllowedRisk), nil
}

func riskRank(risk types.RiskLevel) int {
	switch risk {
	case types.RiskLow:
		return 1
	case types.RiskMedium:
		return 2
	case types.RiskHigh:
		return 3
	case types.RiskCritical:
		return 4
	default:
		return 0
	}
}
