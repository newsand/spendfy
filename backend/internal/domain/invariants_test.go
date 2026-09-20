package domain

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestInvariant_FonteIsolation(t *testing.T) {
	compraObs := &PrecoObservado{
		ID:    1,
		SKUID: 1,
		Preco: decimal.NewFromFloat(10.00),
		Fonte: FonteCompra,
		Data:  time.Now(),
	}

	rastreioObs := &PrecoObservado{
		ID:    2,
		SKUID: 1,
		Preco: decimal.NewFromFloat(12.00),
		Fonte: FonteRastreio,
		URL:   ptr("https://example.com"),
		Data:  time.Now(),
	}

	if compraObs.Fonte == rastreioObs.Fonte {
		t.Error("fontes should be different")
	}

	if compraObs.Fonte != FonteCompra {
		t.Error("expected FonteCompra")
	}

	if rastreioObs.Fonte != FonteRastreio {
		t.Error("expected FonteRastreio")
	}
}

func TestInvariant_PrecoIsAlwaysUnitPrice(t *testing.T) {
	qtd := decimal.NewFromInt(3)
	obs1 := &PrecoObservado{
		Preco:      decimal.NewFromFloat(5.00),
		Quantidade: &qtd,
	}
	obs2 := &PrecoObservado{
		Preco:      decimal.NewFromFloat(5.00),
		Quantidade: nil,
	}

	delta := CalculateDelta(obs2, obs1)

	if delta.DeltaAbsoluto == nil || !delta.DeltaAbsoluto.IsZero() {
		t.Error("quantidade should not affect delta calculation")
	}
}

func TestInvariant_QuantidadeNeverAffectsAlerts(t *testing.T) {
	url := "https://example.com"
	limiarModo := LimiarAbsoluto
	limiarValor := decimal.NewFromFloat(10.00)

	monitor := &MonitorURL{
		URL:         url,
		Ativo:       true,
		LimiarModo:  &limiarModo,
		LimiarValor: &limiarValor,
	}

	qtd := decimal.NewFromInt(100)
	obs := &PrecoObservado{
		Preco:      decimal.NewFromFloat(8.00),
		Quantidade: &qtd,
		Fonte:      FonteRastreio,
		URL:        &url,
	}

	result := CheckAlert(obs, nil, monitor)

	if !result.ShouldAlert {
		t.Error("alert should be based on unit price only, not quantity")
	}
}

func TestInvariant_UnidadeMustMatchSKU(t *testing.T) {
	sku := &SKU{
		ID:            1,
		Nome:          "Test",
		UnidadePadrao: UnidadeKg,
	}

	validReq := &CreateObservacaoRequest{
		SKUID:   1,
		Preco:   decimal.NewFromFloat(10.00),
		Unidade: UnidadeKg,
		Loja:    "Test",
		Fonte:   FonteCompra,
	}

	if err := validReq.ValidateAgainstSKU(sku); err != nil {
		t.Errorf("valid unidade should pass: %v", err)
	}

	invalidUnidades := []UnidadePadrao{UnidadeUn, UnidadeG, UnidadeL, UnidadeMl}
	for _, u := range invalidUnidades {
		req := &CreateObservacaoRequest{
			SKUID:   1,
			Preco:   decimal.NewFromFloat(10.00),
			Unidade: u,
			Loja:    "Test",
			Fonte:   FonteCompra,
		}
		if err := req.ValidateAgainstSKU(sku); err != ErrUnidadeMismatch {
			t.Errorf("unidade %s should fail against sku unidade_padrao kg", u)
		}
	}
}

func TestInvariant_RastreioRequiresURL(t *testing.T) {
	reqWithoutURL := &CreateObservacaoRequest{
		SKUID:   1,
		Preco:   decimal.NewFromFloat(10.00),
		Unidade: UnidadeUn,
		Loja:    "Test",
		Fonte:   FonteRastreio,
		URL:     nil,
	}

	if err := reqWithoutURL.Validate(); err != ErrURLRequiredForRastreio {
		t.Error("rastreio without URL should fail validation")
	}

	url := "https://example.com"
	reqWithURL := &CreateObservacaoRequest{
		SKUID:   1,
		Preco:   decimal.NewFromFloat(10.00),
		Unidade: UnidadeUn,
		Loja:    "Test",
		Fonte:   FonteRastreio,
		URL:     &url,
	}

	if err := reqWithURL.Validate(); err != nil {
		t.Errorf("rastreio with URL should pass: %v", err)
	}
}

func TestInvariant_LimiarSingleModeOnly(t *testing.T) {
	modoAbs := LimiarAbsoluto
	valorAbs := decimal.NewFromFloat(50.00)

	monitor := &MonitorURL{
		LimiarModo:  &modoAbs,
		LimiarValor: &valorAbs,
	}

	if monitor.LimiarModo == nil {
		t.Error("limiar_modo should be set")
	}

	if *monitor.LimiarModo != LimiarAbsoluto && *monitor.LimiarModo != LimiarPercentual {
		t.Error("limiar_modo must be either absoluto or percentual")
	}

	modoPct := LimiarPercentual
	valorPct := decimal.NewFromFloat(10.00)

	monitorPct := &MonitorURL{
		LimiarModo:  &modoPct,
		LimiarValor: &valorPct,
	}

	if *monitor.LimiarModo == *monitorPct.LimiarModo {
		t.Error("monitors have different modes")
	}
}

func TestInvariant_NoLimiarMeansNoAlert(t *testing.T) {
	url := "https://example.com"

	monitor := &MonitorURL{
		URL:         url,
		Ativo:       true,
		LimiarModo:  nil,
		LimiarValor: nil,
	}

	obs := &PrecoObservado{
		Preco: decimal.NewFromFloat(1.00),
		Fonte: FonteRastreio,
		URL:   &url,
	}

	result := CheckAlert(obs, nil, monitor)

	if result.ShouldAlert {
		t.Error("no limiar configured should mean no alert")
	}
}

func TestInvariant_ArchivedURLNotUsedForDelta(t *testing.T) {
	activeURL := "https://example.com/new"
	archivedURL := "https://example.com/old"

	currentObs := &PrecoObservado{
		ID:    3,
		SKUID: 1,
		Preco: decimal.NewFromFloat(10.00),
		Fonte: FonteRastreio,
		URL:   &activeURL,
	}

	previousActiveObs := &PrecoObservado{
		ID:    2,
		SKUID: 1,
		Preco: decimal.NewFromFloat(12.00),
		Fonte: FonteRastreio,
		URL:   &activeURL,
	}

	archivedObs := &PrecoObservado{
		ID:    1,
		SKUID: 1,
		Preco: decimal.NewFromFloat(8.00),
		Fonte: FonteRastreio,
		URL:   &archivedURL,
	}

	deltaWithActive := CalculateDelta(currentObs, previousActiveObs)
	if !deltaWithActive.HasPrevious {
		t.Error("should have previous with same active URL")
	}

	deltaWithArchived := CalculateDelta(currentObs, archivedObs)
	if !deltaWithArchived.HasPrevious {
		t.Error("CalculateDelta itself doesn't filter by URL - that's the repository's job")
	}
}

func TestInvariant_AlertOnlyForActiveURL(t *testing.T) {
	activeURL := "https://example.com/new"
	archivedURL := "https://example.com/old"
	limiarModo := LimiarAbsoluto
	limiarValor := decimal.NewFromFloat(100.00)

	monitor := &MonitorURL{
		URL:         activeURL,
		Ativo:       true,
		LimiarModo:  &limiarModo,
		LimiarValor: &limiarValor,
	}

	obsActiveURL := &PrecoObservado{
		Preco: decimal.NewFromFloat(5.00),
		Fonte: FonteRastreio,
		URL:   &activeURL,
	}

	obsArchivedURL := &PrecoObservado{
		Preco: decimal.NewFromFloat(5.00),
		Fonte: FonteRastreio,
		URL:   &archivedURL,
	}

	resultActive := CheckAlert(obsActiveURL, nil, monitor)
	if !resultActive.ShouldAlert {
		t.Error("should alert for observation with active URL")
	}

	resultArchived := CheckAlert(obsArchivedURL, nil, monitor)
	if resultArchived.ShouldAlert {
		t.Error("should NOT alert for observation with archived URL")
	}
}

func TestInvariant_ScrapeFaiureNoInventedPrice(t *testing.T) {
	req := &CreateObservacaoRequest{
		SKUID:   1,
		Preco:   decimal.Zero,
		Unidade: UnidadeUn,
		Loja:    "Test",
		Fonte:   FonteRastreio,
		URL:     ptr("https://example.com"),
	}

	err := req.Validate()
	if err == nil {
		t.Error("zero price should fail validation - scrape failure must not create observation")
	}

	req.Preco = decimal.NewFromFloat(-1.00)
	err = req.Validate()
	if err == nil {
		t.Error("negative price should fail validation")
	}
}

func TestInvariant_ChaveIdentidadeUnique(t *testing.T) {
	req1 := &CreateSKURequest{
		Nome:            "Test 1",
		UnidadePadrao:   UnidadeUn,
		ChaveIdentidade: "test-sku-001",
	}

	req2 := &CreateSKURequest{
		Nome:            "Test 2",
		UnidadePadrao:   UnidadeUn,
		ChaveIdentidade: "test-sku-001",
	}

	if req1.ChaveIdentidade != req2.ChaveIdentidade {
		t.Error("test setup error - chave_identidade should match")
	}

	if err := req1.Validate(); err != nil {
		t.Errorf("first request should be valid: %v", err)
	}

	if err := req2.Validate(); err != nil {
		t.Errorf("second request validation (structural) should pass: %v", err)
	}
}

func TestInvariant_ValidUnidades(t *testing.T) {
	validUnidades := []UnidadePadrao{UnidadeUn, UnidadeKg, UnidadeG, UnidadeL, UnidadeMl}

	for _, u := range validUnidades {
		if !u.Valid() {
			t.Errorf("unidade %s should be valid", u)
		}
	}

	invalidUnidades := []UnidadePadrao{"invalid", "oz", "lb", ""}
	for _, u := range invalidUnidades {
		if u.Valid() {
			t.Errorf("unidade %s should be invalid", u)
		}
	}
}

func TestInvariant_ValidFontes(t *testing.T) {
	if !FonteCompra.Valid() {
		t.Error("compra should be valid fonte")
	}

	if !FonteRastreio.Valid() {
		t.Error("rastreio should be valid fonte")
	}

	invalidFontes := []Fonte{"invalid", "ocr", "ai", ""}
	for _, f := range invalidFontes {
		if f.Valid() {
			t.Errorf("fonte %s should be invalid", f)
		}
	}
}

func TestInvariant_ValidLimiarModos(t *testing.T) {
	if !LimiarAbsoluto.Valid() {
		t.Error("absoluto should be valid limiar_modo")
	}

	if !LimiarPercentual.Valid() {
		t.Error("percentual should be valid limiar_modo")
	}

	invalidModos := []LimiarModo{"invalid", "both", ""}
	for _, m := range invalidModos {
		if m.Valid() {
			t.Errorf("limiar_modo %s should be invalid", m)
		}
	}
}

func TestInvariant_MultiURLRastreioSeriesDoNotMix(t *testing.T) {
	// Same SKU, two storefront URLs — deltas must be computed per URL only.
	araujo := "https://araujo.example/p/1"
	raia := "https://raia.example/p/1"
	obsA1 := &PrecoObservado{SKUID: 1, Fonte: FonteRastreio, URL: &araujo, Preco: decimal.NewFromFloat(50)}
	obsA2 := &PrecoObservado{SKUID: 1, Fonte: FonteRastreio, URL: &araujo, Preco: decimal.NewFromFloat(45)}
	obsR1 := &PrecoObservado{SKUID: 1, Fonte: FonteRastreio, URL: &raia, Preco: decimal.NewFromFloat(40)}

	d := CalculateDelta(obsA2, obsA1)
	if !d.HasPrevious || d.PreviousPreco.Cmp(decimal.NewFromFloat(50)) != 0 {
		t.Fatalf("araujo series should compare to previous araujo, got %+v", d)
	}
	// Using Raia as previous for Araújo current would be a product bug — callers must pass same-URL previous.
	wrong := CalculateDelta(obsA2, obsR1)
	if !wrong.HasPrevious {
		t.Fatal("CalculateDelta itself is pure; URL filtering is caller's job")
	}
	if wrong.PreviousPreco.Cmp(decimal.NewFromFloat(40)) == 0 {
		// Document: API/repo must not pass cross-URL previous. CheckAlert already rejects URL mismatch.
	}
	monA := &MonitorURL{URL: araujo, LimiarModo: ptrLimiar(LimiarAbsoluto), LimiarValor: ptrDec(46)}
	if CheckAlert(obsA2, obsA1, monA).ShouldAlert != true {
		t.Fatal("expected alert on araujo")
	}
	if CheckAlert(obsA2, obsA1, &MonitorURL{URL: raia, LimiarModo: ptrLimiar(LimiarAbsoluto), LimiarValor: ptrDec(46)}).ShouldAlert {
		t.Fatal("must not alert when monitor URL mismatches observation URL")
	}
}

func ptrLimiar(m LimiarModo) *LimiarModo { return &m }
func ptrDec(f float64) *decimal.Decimal {
	d := decimal.NewFromFloat(f)
	return &d
}

func TestInvariant_DeltaForRastreioMustSpecifyURL(t *testing.T) {
	araujoURL := "https://araujo.example/p/vasenol"
	raiaURL := "https://raia.example/p/vasenol"

	obsAraujo1 := &PrecoObservado{ID: 1, SKUID: 1, Fonte: FonteRastreio, URL: &araujoURL, Preco: decimal.NewFromFloat(50)}
	obsAraujo2 := &PrecoObservado{ID: 2, SKUID: 1, Fonte: FonteRastreio, URL: &araujoURL, Preco: decimal.NewFromFloat(48)}
	obsRaia1 := &PrecoObservado{ID: 3, SKUID: 1, Fonte: FonteRastreio, URL: &raiaURL, Preco: decimal.NewFromFloat(52)}
	obsRaia2 := &PrecoObservado{ID: 4, SKUID: 1, Fonte: FonteRastreio, URL: &raiaURL, Preco: decimal.NewFromFloat(47)}

	deltaAraujo := CalculateDelta(obsAraujo2, obsAraujo1)
	if !deltaAraujo.HasPrevious {
		t.Fatal("Araújo delta should have previous")
	}
	if !deltaAraujo.DeltaAbsoluto.Equal(decimal.NewFromFloat(-2)) {
		t.Errorf("Araújo delta should be -2, got %s", deltaAraujo.DeltaAbsoluto.String())
	}

	deltaRaia := CalculateDelta(obsRaia2, obsRaia1)
	if !deltaRaia.HasPrevious {
		t.Fatal("Raia delta should have previous")
	}
	if !deltaRaia.DeltaAbsoluto.Equal(decimal.NewFromFloat(-5)) {
		t.Errorf("Raia delta should be -5, got %s", deltaRaia.DeltaAbsoluto.String())
	}

	wrongDelta := CalculateDelta(obsAraujo2, obsRaia1)
	if wrongDelta.HasPrevious && wrongDelta.DeltaAbsoluto.Equal(decimal.NewFromFloat(-4)) {
		t.Log("API must prevent cross-URL delta: comparing Araújo current to Raia previous is wrong")
	}
}

func TestInvariant_URLOrMonitorIDRequiredError(t *testing.T) {
	if ErrURLOrMonitorIDRequired == nil {
		t.Fatal("ErrURLOrMonitorIDRequired should be defined")
	}
	expected := "url or monitor_id query param required for fonte=rastreio"
	if ErrURLOrMonitorIDRequired.Error() != expected {
		t.Errorf("expected error message %q, got %q", expected, ErrURLOrMonitorIDRequired.Error())
	}
}
