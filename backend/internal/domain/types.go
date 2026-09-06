package domain

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

type UnidadePadrao string

const (
	UnidadeUn UnidadePadrao = "un"
	UnidadeKg UnidadePadrao = "kg"
	UnidadeG  UnidadePadrao = "g"
	UnidadeL  UnidadePadrao = "l"
	UnidadeMl UnidadePadrao = "ml"
)

func (u UnidadePadrao) Valid() bool {
	switch u {
	case UnidadeUn, UnidadeKg, UnidadeG, UnidadeL, UnidadeMl:
		return true
	}
	return false
}

type Fonte string

const (
	FonteCompra   Fonte = "compra"
	FonteRastreio Fonte = "rastreio"
)

func (f Fonte) Valid() bool {
	return f == FonteCompra || f == FonteRastreio
}

type LimiarModo string

const (
	LimiarAbsoluto   LimiarModo = "absoluto"
	LimiarPercentual LimiarModo = "percentual"
)

func (l LimiarModo) Valid() bool {
	return l == LimiarAbsoluto || l == LimiarPercentual
}

type SKU struct {
	ID               int64         `json:"id"`
	Nome             string        `json:"nome"`
	Marca            *string       `json:"marca,omitempty"`
	TamanhoVariante  *string       `json:"tamanho_ou_variante,omitempty"`
	UnidadePadrao    UnidadePadrao `json:"unidade_padrao"`
	ChaveIdentidade  string        `json:"chave_identidade"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

type PrecoObservado struct {
	ID         int64           `json:"id"`
	SKUID      int64           `json:"sku_id"`
	Preco      decimal.Decimal `json:"preco"`
	Moeda      string          `json:"moeda"`
	Quantidade *decimal.Decimal `json:"quantidade,omitempty"`
	Unidade    UnidadePadrao   `json:"unidade"`
	Loja       string          `json:"loja"`
	Data       time.Time       `json:"data"`
	Fonte      Fonte           `json:"fonte"`
	URL        *string         `json:"url,omitempty"`
	Notas      *string         `json:"notas,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type MonitorURL struct {
	ID           int64           `json:"id"`
	SKUID        int64           `json:"sku_id"`
	URL          string          `json:"url"`
	Ativo        bool            `json:"ativo"`
	LimiarModo   *LimiarModo     `json:"limiar_modo,omitempty"`
	LimiarValor  *decimal.Decimal `json:"limiar_valor,omitempty"`
	CSSSelector  *string         `json:"css_selector,omitempty"`
	RegexPattern *string         `json:"regex_pattern,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	ArchivedAt   *time.Time      `json:"archived_at,omitempty"`
}

var (
	ErrSKUNotFound            = errors.New("sku not found")
	ErrObservacaoNotFound     = errors.New("observacao not found")
	ErrMonitorNotFound        = errors.New("monitor not found")
	ErrUnidadeMismatch        = errors.New("unidade must match sku unidade_padrao")
	ErrURLRequiredForRastreio = errors.New("url required for fonte=rastreio")
	ErrChaveIdentidadeExists  = errors.New("chave_identidade already exists")
	ErrInvalidUnidade         = errors.New("invalid unidade")
	ErrInvalidFonte           = errors.New("invalid fonte")
	ErrInvalidLimiarModo      = errors.New("invalid limiar_modo")
	ErrLimiarModoRequired     = errors.New("limiar_modo and limiar_valor must both be set or both be null")
)

type TrackedSKU struct {
	SKU                SKU                 `json:"sku"`
	LastCompra         *decimal.Decimal    `json:"last_compra,omitempty"`
	LastCompraData     *time.Time          `json:"last_compra_data,omitempty"`
	HasMonitor         bool                `json:"has_monitor"`
	SeriesCompra       []PricePoint        `json:"series_compra"`
	SeriesRastreioByURL []RastreioSeries   `json:"series_rastreio_by_url"`
}

type PricePoint struct {
	Data  time.Time       `json:"data"`
	Preco decimal.Decimal `json:"preco"`
}

type RastreioSeries struct {
	URL    string       `json:"url"`
	Points []PricePoint `json:"points"`
}
