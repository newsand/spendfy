package domain

import "github.com/shopspring/decimal"

type DeltaResult struct {
	CurrentPreco    decimal.Decimal  `json:"current_preco"`
	PreviousPreco   *decimal.Decimal `json:"previous_preco,omitempty"`
	DeltaAbsoluto   *decimal.Decimal `json:"delta_absoluto,omitempty"`
	DeltaPercentual *decimal.Decimal `json:"delta_percentual,omitempty"`
	HasPrevious     bool             `json:"has_previous"`
}

func CalculateDelta(current, previous *PrecoObservado) DeltaResult {
	result := DeltaResult{
		CurrentPreco: current.Preco,
		HasPrevious:  previous != nil,
	}

	if previous == nil {
		return result
	}

	result.PreviousPreco = &previous.Preco

	deltaAbs := current.Preco.Sub(previous.Preco)
	result.DeltaAbsoluto = &deltaAbs

	if !previous.Preco.IsZero() {
		deltaPct := deltaAbs.Div(previous.Preco).Mul(decimal.NewFromInt(100))
		result.DeltaPercentual = &deltaPct
	}

	return result
}

type AlertCheck struct {
	ShouldAlert bool
	Reason      string
}

func CheckAlert(obs *PrecoObservado, previous *PrecoObservado, monitor *MonitorURL) AlertCheck {
	if monitor == nil || monitor.LimiarModo == nil || monitor.LimiarValor == nil {
		return AlertCheck{ShouldAlert: false, Reason: "no limiar configured"}
	}

	if obs.Fonte != FonteRastreio {
		return AlertCheck{ShouldAlert: false, Reason: "alerts only for rastreio"}
	}

	if obs.URL == nil || *obs.URL != monitor.URL {
		return AlertCheck{ShouldAlert: false, Reason: "url mismatch with active monitor"}
	}

	switch *monitor.LimiarModo {
	case LimiarAbsoluto:
		if obs.Preco.LessThanOrEqual(*monitor.LimiarValor) {
			return AlertCheck{
				ShouldAlert: true,
				Reason:      "preco <= limiar_valor (absoluto)",
			}
		}
	case LimiarPercentual:
		if previous == nil {
			return AlertCheck{ShouldAlert: false, Reason: "no previous observation for percentual check"}
		}
		if previous.Preco.IsZero() {
			return AlertCheck{ShouldAlert: false, Reason: "previous preco is zero"}
		}
		drop := previous.Preco.Sub(obs.Preco).Div(previous.Preco).Mul(decimal.NewFromInt(100))
		if drop.GreaterThanOrEqual(*monitor.LimiarValor) {
			return AlertCheck{
				ShouldAlert: true,
				Reason:      "price drop >= limiar_valor (percentual)",
			}
		}
	}

	return AlertCheck{ShouldAlert: false, Reason: "limiar not met"}
}
