package domain

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCreateObservacaoRequest_UnidadeMismatch(t *testing.T) {
	sku := &SKU{
		ID:            1,
		UnidadePadrao: UnidadeKg,
	}

	req := &CreateObservacaoRequest{
		SKUID:   1,
		Preco:   decimal.NewFromFloat(10.00),
		Unidade: UnidadeUn,
		Loja:    "Test",
		Fonte:   FonteCompra,
	}

	err := req.ValidateAgainstSKU(sku)
	if err != ErrUnidadeMismatch {
		t.Errorf("expected ErrUnidadeMismatch, got %v", err)
	}
}

func TestCreateObservacaoRequest_UnidadeMatch(t *testing.T) {
	sku := &SKU{
		ID:            1,
		UnidadePadrao: UnidadeKg,
	}

	req := &CreateObservacaoRequest{
		SKUID:   1,
		Preco:   decimal.NewFromFloat(10.00),
		Unidade: UnidadeKg,
		Loja:    "Test",
		Fonte:   FonteCompra,
	}

	err := req.ValidateAgainstSKU(sku)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCreateObservacaoRequest_RastreioRequiresURL(t *testing.T) {
	req := &CreateObservacaoRequest{
		SKUID:   1,
		Preco:   decimal.NewFromFloat(10.00),
		Unidade: UnidadeUn,
		Loja:    "Test",
		Fonte:   FonteRastreio,
		URL:     nil,
	}

	err := req.Validate()
	if err != ErrURLRequiredForRastreio {
		t.Errorf("expected ErrURLRequiredForRastreio, got %v", err)
	}
}

func TestCreateObservacaoRequest_CompraNoURLRequired(t *testing.T) {
	req := &CreateObservacaoRequest{
		SKUID:   1,
		Preco:   decimal.NewFromFloat(10.00),
		Unidade: UnidadeUn,
		Loja:    "Test",
		Fonte:   FonteCompra,
		URL:     nil,
	}

	err := req.Validate()
	if err != nil {
		t.Errorf("expected no error for compra without URL, got %v", err)
	}
}

func TestSetMonitorRequest_LimiarBothOrNone(t *testing.T) {
	limiarModo := LimiarAbsoluto

	tests := []struct {
		name        string
		limiarModo  *LimiarModo
		limiarValor *decimal.Decimal
		expectErr   bool
	}{
		{
			name:        "both nil - ok",
			limiarModo:  nil,
			limiarValor: nil,
			expectErr:   false,
		},
		{
			name:        "both set - ok",
			limiarModo:  &limiarModo,
			limiarValor: ptr(decimal.NewFromFloat(10.00)),
			expectErr:   false,
		},
		{
			name:        "only modo - error",
			limiarModo:  &limiarModo,
			limiarValor: nil,
			expectErr:   true,
		},
		{
			name:        "only valor - error",
			limiarModo:  nil,
			limiarValor: ptr(decimal.NewFromFloat(10.00)),
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &SetMonitorRequest{
				SKUID:       1,
				URL:         "https://example.com",
				LimiarModo:  tt.limiarModo,
				LimiarValor: tt.limiarValor,
			}
			err := req.Validate()
			if tt.expectErr && err == nil {
				t.Error("expected error")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
