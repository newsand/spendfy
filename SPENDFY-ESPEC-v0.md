# Spendfy — especificação v0

**Repo:** https://github.com/newsand/spendfy  
**Status:** v0 atacável (job fechado + unidade canônica + fora de escopo sem “etc”)  
**Data:** 2026-09-06

---

## 1. Job do usuário (fechado)

**Pergunta canônica:** para um **SKU** estável, “este item custou mais ou menos que da *última observação do mesmo tipo*?”

Exemplos válidos:
- “Esse pão custa mais ou menos que da última vez que eu registrei a compra?”
- “Esse shampoo (rastreio no site) abaixou de preço em relação à última coleta?”

**Não é o job da v0:**
- “Minha compra / cesta ficou mais cara” (comparação de cesta)
- Inflação agregada / IPCA pessoal
- Orçamento, carteira, categorias de gasto

---

## 2. Unidade canônica

### 2.1 SKU

Registro mínimo de **produto acompanhado**.

| Campo | Obrigatório | Nota |
|-------|-------------|------|
| `id` | sim | interno |
| `nome` | sim | rótulo humano (ex.: “Vasenol Clinical 200ml”) |
| `marca` | não | |
| `tamanho_ou_variante` | não | texto livre na v0 (ex.: “200ml”) |
| `unidade_padrao` | sim | `un` \| `kg` \| `g` \| `l` \| `ml` — unidade em que o preço é comparado |
| `chave_identidade` | sim | ver §2.3 |

Na v0 **não** há matching fuzzy automático entre nomes de notinha/site. O usuário (ou o form) amarra a observação a um SKU existente ou cria um novo.

### 2.2 Preço observado

Registro mínimo de **uma observação de preço** ligada a um SKU.

| Campo | Obrigatório | Nota |
|-------|-------------|------|
| `id` | sim | |
| `sku_id` | sim | FK para SKU |
| `preco` | sim | valor monetário (BRL na v0) |
| `moeda` | sim | fixo `BRL` na v0 |
| `quantidade` | sim | quantidade na unidade da observação (default 1) |
| `unidade` | sim | deve ser comparável à `unidade_padrao` do SKU (mesma unidade ou conversão explícita documentada; na v0: **mesma unidade**, sem conversão) |
| `loja` | sim | nome ou identificador da loja/site |
| `data` | sim | data/hora da observação (UTC armazenado; exibição America/Sao_Paulo) |
| `fonte` | sim | `compra` \| `rastreio` — ver §2.4 |
| `url` | se `fonte=rastreio` | URL monitorada |
| `notas` | não | texto livre |

**Promoção:** fora da v0 como campo estruturado. Se o usuário quiser registrar preço promocional via form, entra como preço observado normal; não há flag `promocao`.

### 2.3 Identidade do SKU (regra fechada)

Dois registros são o **mesmo SKU** só se compartilham a mesma `chave_identidade`.

Na v0, `chave_identidade` é definida **pelo usuário no cadastro** (string estável). Sugestão de preenchimento (não automática):
- compra recorrente: `marca + nome + tamanho` digitados de forma consistente
- rastreio: URL canônica do produto no site (uma URL = um alvo de scrape; o SKU aponta para essa URL no monitor)

**Proibido na v0:** inferir “é o mesmo produto” por similaridade de nome entre Araújo e Raia, ou entre linha de notinha e catálogo.

### 2.4 Fonte e comparação (“última observação”)

`fonte` distingue verdade diferente:

| Valor | Significado | Origem na v0 |
|-------|-------------|--------------|
| `compra` | preço que **você pagou** | form manual |
| `rastreio` | preço que o **site mostrou** | scraper de 1 URL |

**Regra de comparação:** Δ% / Δ$ entre observação atual e a **anterior com o mesmo `sku_id` e a mesma `fonte`**.  
Misturar `compra` com `rastreio` na mesma série é **inválido**. A UI pode mostrar as duas séries lado a lado; não calcula “última vez” cruzando tipos.

---

## 3. Superfícies da v0 (duas entradas magras)

### 3.1 Form manual (`fonte=compra`)

- Criar/editar SKU
- Registrar preço observado (loja, data, unidade, preço) amarrado a um SKU
- Ver histórico do SKU (série `compra`) e Δ vs última `compra`

### 3.2 Monitor de 1 URL (`fonte=rastreio`)

- Por SKU: cadastrar **no máximo uma** URL ativa de monitor
- Job diário: fetch/scrape dessa URL, extrair preço, gravar preço observado `rastreio`
- Se o scrape falhar: **não** inventar preço; registrar falha de coleta (status/log) e manter última observação válida intacta
- Limiar de alerta: por SKU, valor absoluto (`preco <= X`) **ou** queda percentual vs última `rastreio` (`queda_pct >= Y`) — pelo menos um dos dois configurado; ambos explícitos na UI/API, sem default mágico de “barato”
- Notificação: **somente Telegram** (um canal). E-mail e WhatsApp fora.

### 3.3 O que a v0 não promete nestas superfícies

- OCR / upload de notinha fiscal
- Várias URLs ou vários sites por SKU (ex.: 5 farmácias)
- Catálogo multi-canal de notificação
- Matching automático SKU entre lojas

---

## 4. Fora de escopo (explícito)

- OCR / parse de notinha ou cupom fiscal
- Comparação de cesta / “a compra ficou mais cara”
- Multi-site por SKU (N farmácias)
- WhatsApp, e-mail, push, SMS (além do Telegram único)
- Flag estruturada de promoção / preço de atacado vs varejo
- Conversão automática de unidades (kg↔g, ml↔l)
- Auth multi-usuário / times / compartilhamento (v0 = single-user)
- App mobile nativo
- Pagamentos, orçamento, metas, categorias financeiras
- Inflação agregada / índices públicos
- Matching fuzzy / ML de identidade de produto

---

## 5. Critérios de aceite (v0 atacável)

1. Dado um SKU com ≥2 observações `compra`, a UI/API responde Δ vs a observação `compra` imediatamente anterior (mesmo `sku_id`, mesma `fonte`).
2. Dado um SKU com URL de monitor e limiar configurado, uma coleta `rastreio` abaixo do limiar dispara **uma** mensagem Telegram; sem limiar configurado, **não** alerta.
3. Uma falha de scrape não cria preço observado falso.
4. Tentativa de comparar ou alertar misturando `compra` e `rastreio` é rejeitada ou não oferecida.
5. Não existe endpoint/tela de OCR, segundo canal de notificação, ou segunda URL ativa por SKU.

---

## 6. Decisões abertas (não descrever tela/regra até fechar)

Estas ficam **nomeadas** e **sem desenho** na v0 até ADR:

- D1 — Stack (linguagem, DB, hosting do job diário)
- D2 — Como o bot Telegram associa chat_id ao usuário único
- D3 — Seletor CSS / estratégia de extração de preço por URL (por site vs genérico)
- D4 — Retenção de histórico (ilimitado vs janela)

Qualquer tela ou regra sobre D1–D4 antes de fechar a decisão é **inválida** nesta spec.

---

## 7. Próximo passo

1. @Socrates — julgar se o job + unidade + fora batem com “a coisa certa”.
2. @Moriaty — atacar buracos exploráveis nesta doc.
3. Após passe: implementação da fatia v0 (form + 1 monitor + Telegram + limiar).
