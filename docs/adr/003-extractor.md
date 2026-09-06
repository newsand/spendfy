# ADR-003: Extração de preço por seletor

## Status

Accepted (D3 fechada para v0)

## Context

O worker precisa extrair preços de páginas HTML. Sites têm estruturas diferentes, e não queremos IA como fonte de preço (spec v0).

## Decision

Três modos de extração, configuráveis por monitor:

### 1. CSS Selector

Campo `css_selector` no monitor. Worker usa goquery para selecionar elemento e extrair texto.

```
css_selector: ".product-price"
css_selector: "#main-price span.value"
```

### 2. Regex Pattern

Campo `regex_pattern` no monitor. Primeiro grupo capturado é o preço.

```
regex_pattern: "R\\$\\s*([\\d.,]+)"
regex_pattern: "\"price\":\\s*\"?([\\d.,]+)"
```

### 3. Default (fallback)

Se nenhum seletor configurado, tenta padrões comuns em ordem:

1. `R\$\s*([\d.,]+)` — formato brasileiro
2. `"price":\s*"?([\d.,]+)"?` — JSON-LD ou data attributes
3. `data-price="([\d.,]+)"` — data attributes HTML5
4. `class="[^"]*price[^"]*"[^>]*>([\d.,\s]+)` — elementos com classe "price"

## Parsing de valor

Suporta formatos brasileiros e internacionais:

- `1.234,56` → `1234.56`
- `1,234.56` → `1234.56`
- `10,50` → `10.50`

## Consequences

- Usuário precisa descobrir o seletor correto para sites menos comuns
- Falha de extração não inventa preço (spec v0) — registra erro no log
- Configuração é por URL/monitor, não global

## Future considerations

- UI para testar seletores antes de salvar
- Sugestão automática de seletor (sem IA escrevendo preço)
- Histórico de erros de extração visível no front
