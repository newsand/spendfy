# ADR-003: Extração de preço por seletor

## Status

Accepted (D3 fechada para v0)

## Context

O worker precisa extrair preços de páginas HTML. Sites têm estruturas diferentes, e não queremos IA como fonte de preço (spec v0).

**Problema crítico:** Muitas PDPs (Product Detail Pages) mostram múltiplos preços na mesma página:
- Preço unitário principal (o que queremos)
- Preços de kits/combos (ex: "Leve 3 pague R$143,97")
- Preços de produtos relacionados
- Preços de parcelas

Se o seletor for muito genérico (ex: `.price`), pode capturar o preço errado e corromper a série.

## Decision

### Regra fundamental

O seletor CSS ou regex **DEVE** apontar especificamente para o **preço unitário principal** — aquele exibido junto ao botão "Comprar" / "Adicionar ao carrinho". Kits e combos na mesma página são **SKUs diferentes** e devem ter **URLs/monitores diferentes**.

### Três modos de extração

#### 1. CSS Selector (recomendado)

Campo `css_selector` no monitor. Deve ser específico o suficiente para retornar apenas o preço unitário.

**Bom (específico):**
```
.product-main-info .product-price-box .price-value
[data-product-price]
.product-price-box > .price-value:first-child
```

**Ruim (muito genérico):**
```
.price
.price-value
span.price
```

#### 2. Regex Pattern

Campo `regex_pattern` no monitor. Primeiro grupo capturado é o preço.

```
data-product-price="([\d.]+)"
"mainPrice":\s*"?([\d.,]+)"
```

#### 3. Default (DESABILITADO em v0)

**NÃO há fallback automático.** Se `css_selector` e `regex_pattern` estiverem vazios, o worker **falha a coleta** e não posta observação.

Motivo: fallbacks genéricos como `"price":` ou `R\$\s*` podem capturar preço de kit/combo em vez do unitário, corrompendo a série.

Regra v0: **todo monitor DEVE ter css_selector ou regex_pattern configurado.**

### Comportamento em caso de ambiguidade

Se o seletor retornar **múltiplos preços diferentes**, o extrator **DEVE falhar** em vez de adivinhar:

```go
if len(distinctPrices) > 1 {
    return ErrAmbiguousMatch
}
```

Isso garante que um seletor mal configurado não corrompa a série com preços de kit.

**Importante:** O worker usa `ExtractCSSStrict` (não `ExtractCSS`). Se o DOM tiver kit price antes do unit price e o seletor for genérico (`.price-value`), a coleta falha em vez de postar o kit price.

## Parsing de valor

Suporta formato brasileiro:

- `R$ 47,99` → `47.99`
- `R$ 1.234,56` → `1234.56`
- `143,97` → `143.97`

## Exemplo: Araújo PDP

A página de produto da Araújo mostra:
- Preço principal: R$ 47,99 (unidade)
- Kit 3x: R$ 143,97
- Kit 4x: R$ 179,96

Seletor correto para o Vasenol 200ml:
```
.product-main-info .product-price-box .price-value
```

Ou via data attribute:
```
[data-product-price]
```

## Consequences

- Usuário **deve** inspecionar a página e escolher seletor específico
- Falha de extração ou ambiguidade **não** cria observação (spec v0)
- Kits são SKUs diferentes → URLs/monitores diferentes
- Configuração é por URL/monitor, não global

## Testes de regressão

`worker/internal/extractor/extractor_test.go` (unit tests):

- `TestExtractCSS_AraujoUnitPrice` — seletor específico retorna 47.99
- `TestExtractCSS_KitPriceNotReturned` — nunca retorna 143.97/179.96
- `TestExtractCSSStrict_AmbiguousMultiMatch` — seletor genérico falha
- `TestInvariant_ScrapeFailureNoInventedPrice` — sem match = erro

`worker/cmd/collector/collector_test.go` (integration tests):

- `TestCollector_ScrapePriceRejectsBroadSelectorWithKitFirst` — **Bug 1 fix**: DOM com kit price antes do unit price + seletor `.price-value` → erro, não posta 143.97
- `TestCollector_ScrapePriceAcceptsSpecificSelector` — seletor específico funciona
- `TestCollector_ScrapePriceRejectsNoSelector` — **Bug 3 fix**: sem seletor configurado → erro
- `TestCollector_ScrapePriceWithRegexWorks` — regex funciona

## Limitação: WAF / Akamai 403

Algumas PDPs (incluindo Araújo) são protegidas por WAF (Akamai, Cloudflare, etc.) que retornam **HTTP 403** para requests HTTP simples sem headers de browser ou cookies.

Comportamento v0:
- Collector faz `GET` com User-Agent básico
- Se 403 → scrape falha → **não posta observação** (correto por spec)
- Fixture HTML nos testes simula estrutura da página, não o WAF

**Live e2e para targets protegidos requer follow-up:**
- Headers realistas / cookies
- Headless browser (Playwright, Puppeteer)
- Proxy rotativo

Tracked in [#2](https://github.com/newsand/spendfy/issues/2).

Este PR não implementa anti-bot. O collector falha graciosamente.

## Future considerations

- UI para testar seletores antes de salvar
- Preview do preço extraído antes de ativar monitor
- Validação: preço extraído deve estar em range razoável vs histórico
- Browser-based scraping para sites com WAF
