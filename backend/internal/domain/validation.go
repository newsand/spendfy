package domain

import "github.com/shopspring/decimal"

type CreateSKURequest struct {
	Nome            string        `json:"nome"`
	Marca           *string       `json:"marca,omitempty"`
	TamanhoVariante *string       `json:"tamanho_ou_variante,omitempty"`
	UnidadePadrao   UnidadePadrao `json:"unidade_padrao"`
	ChaveIdentidade string        `json:"chave_identidade"`
}

func (r *CreateSKURequest) Validate() error {
	if r.Nome == "" {
		return &ValidationError{Field: "nome", Message: "required"}
	}
	if r.ChaveIdentidade == "" {
		return &ValidationError{Field: "chave_identidade", Message: "required"}
	}
	if !r.UnidadePadrao.Valid() {
		return ErrInvalidUnidade
	}
	return nil
}

type UpdateSKURequest struct {
	Nome            *string        `json:"nome,omitempty"`
	Marca           *string        `json:"marca,omitempty"`
	TamanhoVariante *string        `json:"tamanho_ou_variante,omitempty"`
	UnidadePadrao   *UnidadePadrao `json:"unidade_padrao,omitempty"`
}

func (r *UpdateSKURequest) Validate() error {
	if r.UnidadePadrao != nil && !r.UnidadePadrao.Valid() {
		return ErrInvalidUnidade
	}
	return nil
}

type CreateObservacaoRequest struct {
	SKUID      int64            `json:"sku_id"`
	Preco      decimal.Decimal  `json:"preco"`
	Quantidade *decimal.Decimal `json:"quantidade,omitempty"`
	Unidade    UnidadePadrao    `json:"unidade"`
	Loja       string           `json:"loja"`
	Fonte      Fonte            `json:"fonte"`
	URL        *string          `json:"url,omitempty"`
	Notas      *string          `json:"notas,omitempty"`
}

func (r *CreateObservacaoRequest) Validate() error {
	if r.SKUID <= 0 {
		return &ValidationError{Field: "sku_id", Message: "required"}
	}
	if r.Preco.IsNegative() || r.Preco.IsZero() {
		return &ValidationError{Field: "preco", Message: "must be positive"}
	}
	if !r.Unidade.Valid() {
		return ErrInvalidUnidade
	}
	if r.Loja == "" {
		return &ValidationError{Field: "loja", Message: "required"}
	}
	if !r.Fonte.Valid() {
		return ErrInvalidFonte
	}
	if r.Fonte == FonteRastreio && (r.URL == nil || *r.URL == "") {
		return ErrURLRequiredForRastreio
	}
	return nil
}

func (r *CreateObservacaoRequest) ValidateAgainstSKU(sku *SKU) error {
	if r.Unidade != sku.UnidadePadrao {
		return ErrUnidadeMismatch
	}
	return nil
}

type SetMonitorRequest struct {
	SKUID        int64            `json:"sku_id"`
	URL          string           `json:"url"`
	LimiarModo   *LimiarModo      `json:"limiar_modo,omitempty"`
	LimiarValor  *decimal.Decimal `json:"limiar_valor,omitempty"`
	CSSSelector  *string          `json:"css_selector,omitempty"`
	RegexPattern *string          `json:"regex_pattern,omitempty"`
}

func (r *SetMonitorRequest) Validate() error {
	if r.SKUID <= 0 {
		return &ValidationError{Field: "sku_id", Message: "required"}
	}
	if r.URL == "" {
		return &ValidationError{Field: "url", Message: "required"}
	}
	hasLimiarModo := r.LimiarModo != nil
	hasLimiarValor := r.LimiarValor != nil
	if hasLimiarModo != hasLimiarValor {
		return ErrLimiarModoRequired
	}
	if r.LimiarModo != nil && !r.LimiarModo.Valid() {
		return ErrInvalidLimiarModo
	}
	return nil
}

type UpdateMonitorLimiarRequest struct {
	MonitorID   int64        `json:"monitor_id,omitempty"`
	LimiarModo  *LimiarModo      `json:"limiar_modo"`
	LimiarValor *decimal.Decimal `json:"limiar_valor"`
}

func (r *UpdateMonitorLimiarRequest) Validate() error {
	hasLimiarModo := r.LimiarModo != nil
	hasLimiarValor := r.LimiarValor != nil
	if hasLimiarModo != hasLimiarValor {
		return ErrLimiarModoRequired
	}
	if r.LimiarModo != nil && !r.LimiarModo.Valid() {
		return ErrInvalidLimiarModo
	}
	return nil
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
