package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func ptr[T any](v T) *T {
	return &v
}

func TestCalculateDelta_NoPrevious(t *testing.T) {
	current := &PrecoObservado{
		ID:    1,
		SKUID: 1,
		Preco: decimal.NewFromFloat(10.50),
		Fonte: FonteCompra,
	}

	result := CalculateDelta(current, nil)

	if result.HasPrevious {
		t.Error("expected HasPrevious to be false")
	}
	if result.PreviousPreco != nil {
		t.Error("expected PreviousPreco to be nil")
	}
	if result.DeltaAbsoluto != nil {
		t.Error("expected DeltaAbsoluto to be nil")
	}
}

func TestCalculateDelta_WithPrevious(t *testing.T) {
	current := &PrecoObservado{
		ID:    2,
		SKUID: 1,
		Preco: decimal.NewFromFloat(12.00),
		Fonte: FonteCompra,
	}
	previous := &PrecoObservado{
		ID:    1,
		SKUID: 1,
		Preco: decimal.NewFromFloat(10.00),
		Fonte: FonteCompra,
	}

	result := CalculateDelta(current, previous)

	if !result.HasPrevious {
		t.Error("expected HasPrevious to be true")
	}
	if result.DeltaAbsoluto == nil {
		t.Fatal("expected DeltaAbsoluto to be set")
	}
	if !result.DeltaAbsoluto.Equal(decimal.NewFromFloat(2.00)) {
		t.Errorf("expected delta absoluto 2.00, got %s", result.DeltaAbsoluto.String())
	}
	if result.DeltaPercentual == nil {
		t.Fatal("expected DeltaPercentual to be set")
	}
	if !result.DeltaPercentual.Equal(decimal.NewFromFloat(20.00)) {
		t.Errorf("expected delta percentual 20.00, got %s", result.DeltaPercentual.String())
	}
}

func TestCalculateDelta_QuantidadeNotUsed(t *testing.T) {
	qtd1 := decimal.NewFromInt(2)
	qtd2 := decimal.NewFromInt(5)

	current := &PrecoObservado{
		ID:         2,
		SKUID:      1,
		Preco:      decimal.NewFromFloat(10.00),
		Quantidade: &qtd1,
		Fonte:      FonteCompra,
	}
	previous := &PrecoObservado{
		ID:         1,
		SKUID:      1,
		Preco:      decimal.NewFromFloat(10.00),
		Quantidade: &qtd2,
		Fonte:      FonteCompra,
	}

	result := CalculateDelta(current, previous)

	if !result.DeltaAbsoluto.IsZero() {
		t.Errorf("quantidade should not affect delta, got %s", result.DeltaAbsoluto.String())
	}
}

func TestCheckAlert_NoLimiarConfigured(t *testing.T) {
	obs := &PrecoObservado{
		Preco: decimal.NewFromFloat(5.00),
		Fonte: FonteRastreio,
		URL:   ptr("https://example.com/product"),
	}
	monitor := &MonitorURL{
		URL:   "https://example.com/product",
		Ativo: true,
	}

	result := CheckAlert(obs, nil, monitor)

	if result.ShouldAlert {
		t.Error("should not alert when no limiar configured")
	}
}

func TestCheckAlert_AbsolutoMode(t *testing.T) {
	url := "https://example.com/product"
	limiarModo := LimiarAbsoluto
	limiarValor := decimal.NewFromFloat(8.00)

	monitor := &MonitorURL{
		URL:         url,
		Ativo:       true,
		LimiarModo:  &limiarModo,
		LimiarValor: &limiarValor,
	}

	tests := []struct {
		name        string
		preco       decimal.Decimal
		shouldAlert bool
	}{
		{"price below limiar", decimal.NewFromFloat(7.00), true},
		{"price equals limiar", decimal.NewFromFloat(8.00), true},
		{"price above limiar", decimal.NewFromFloat(9.00), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := &PrecoObservado{
				Preco: tt.preco,
				Fonte: FonteRastreio,
				URL:   &url,
			}
			result := CheckAlert(obs, nil, monitor)
			if result.ShouldAlert != tt.shouldAlert {
				t.Errorf("expected shouldAlert=%v, got %v", tt.shouldAlert, result.ShouldAlert)
			}
		})
	}
}

func TestCheckAlert_PercentualMode(t *testing.T) {
	url := "https://example.com/product"
	limiarModo := LimiarPercentual
	limiarValor := decimal.NewFromFloat(10.00)

	monitor := &MonitorURL{
		URL:         url,
		Ativo:       true,
		LimiarModo:  &limiarModo,
		LimiarValor: &limiarValor,
	}

	previous := &PrecoObservado{
		Preco: decimal.NewFromFloat(100.00),
		Fonte: FonteRastreio,
		URL:   &url,
		Data:  time.Now().Add(-24 * time.Hour),
	}

	tests := []struct {
		name        string
		preco       decimal.Decimal
		shouldAlert bool
	}{
		{"drop 15%", decimal.NewFromFloat(85.00), true},
		{"drop exactly 10%", decimal.NewFromFloat(90.00), true},
		{"drop 5%", decimal.NewFromFloat(95.00), false},
		{"price increase", decimal.NewFromFloat(110.00), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := &PrecoObservado{
				Preco: tt.preco,
				Fonte: FonteRastreio,
				URL:   &url,
			}
			result := CheckAlert(obs, previous, monitor)
			if result.ShouldAlert != tt.shouldAlert {
				t.Errorf("expected shouldAlert=%v, got %v (reason: %s)", tt.shouldAlert, result.ShouldAlert, result.Reason)
			}
		})
	}
}

func TestCheckAlert_OnlyRastreio(t *testing.T) {
	url := "https://example.com/product"
	limiarModo := LimiarAbsoluto
	limiarValor := decimal.NewFromFloat(100.00)

	obs := &PrecoObservado{
		Preco: decimal.NewFromFloat(5.00),
		Fonte: FonteCompra,
		URL:   &url,
	}
	monitor := &MonitorURL{
		URL:         url,
		Ativo:       true,
		LimiarModo:  &limiarModo,
		LimiarValor: &limiarValor,
	}

	result := CheckAlert(obs, nil, monitor)

	if result.ShouldAlert {
		t.Error("should not alert for fonte=compra")
	}
}

func TestCheckAlert_URLMismatch(t *testing.T) {
	activeURL := "https://example.com/product-new"
	obsURL := "https://example.com/product-old"
	limiarModo := LimiarAbsoluto
	limiarValor := decimal.NewFromFloat(100.00)

	obs := &PrecoObservado{
		Preco: decimal.NewFromFloat(5.00),
		Fonte: FonteRastreio,
		URL:   &obsURL,
	}
	monitor := &MonitorURL{
		URL:         activeURL,
		Ativo:       true,
		LimiarModo:  &limiarModo,
		LimiarValor: &limiarValor,
	}

	result := CheckAlert(obs, nil, monitor)

	if result.ShouldAlert {
		t.Error("should not alert when URL doesn't match active monitor")
	}
}

func TestCheckAlert_NeverBothModes(t *testing.T) {
	url := "https://example.com/product"

	modoAbs := LimiarAbsoluto
	valorAbs := decimal.NewFromFloat(50.00)

	monitorAbs := &MonitorURL{
		URL:         url,
		Ativo:       true,
		LimiarModo:  &modoAbs,
		LimiarValor: &valorAbs,
	}

	obs := &PrecoObservado{
		Preco: decimal.NewFromFloat(45.00),
		Fonte: FonteRastreio,
		URL:   &url,
	}

	previous := &PrecoObservado{
		Preco: decimal.NewFromFloat(100.00),
		Fonte: FonteRastreio,
		URL:   &url,
	}

	result := CheckAlert(obs, previous, monitorAbs)
	if !result.ShouldAlert {
		t.Error("should alert for absoluto mode")
	}
	if result.Reason != "preco <= limiar_valor (absoluto)" {
		t.Errorf("should use absoluto reason, got: %s", result.Reason)
	}
}
