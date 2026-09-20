# Spendfy — especificação v0

**Repo:** https://github.com/newsand/spendfy  
**Status:** v0 atacável (patches pós-ataque Moriaty/Socrates)  
**Data:** 2026-09-06  
**Rev:** 0.4

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
| `preco` | sim | **sempre preço unitário** na `unidade` (BRL). Nunca total da linha. |
| `moeda` | sim | fixo `BRL` na v0 |
| `quantidade` | não | só contexto de compra (“comprei 2”); **não entra** no Δ nem no limiar |
| `unidade` | sim | na v0: **obrigatoriamente igual** a `unidade_padrao` do SKU (sem conversão) |
| `loja` | sim | nome ou identificador da loja/site |
| `data` | sim | data/hora da observação (UTC armazenado; exibição America/Sao_Paulo) |
| `fonte` | sim | `compra` \| `rastreio` — ver §2.4 |
| `url` | se `fonte=rastreio` | URL da qual veio esta observação |
| `notas` | não | texto livre |

**Regra fechada (`preco` × `quantidade`):** a série e os alertas usam **somente** `preco` unitário. Se o usuário pagou R$10 por 2 un, o form grava `preco=5`, `unidade=un`, `quantidade=2` (opcional). Gravar `preco=10` com `quantidade=2` como se fosse total é **inválido**.

**Promoção:** fora da v0 como campo estruturado.

### 2.3 Identidade do SKU (regra fechada)

Dois registros são o **mesmo SKU** só se compartilham a mesma `chave_identidade`.

Na v0, `chave_identidade` é definida **pelo usuário no cadastro** (string estável). Sugestão de preenchimento (não automática):
- compra recorrente: `marca + nome + tamanho` digitados de forma consistente
- rastreio: amarra ao monitor ativo (§3.2); identidade do SKU **não** é inferida da URL automaticamente além do vínculo explícito

**Proibido na v0:** inferir “é o mesmo produto” por similaridade de nome entre lojas, ou entre linha de notinha e catálogo.

### 2.4 Fonte e comparação (“última observação”)

`fonte` distingue verdade diferente:

| Valor | Significado | Origem na v0 |
|-------|-------------|--------------|
| `compra` | preço unitário que **você pagou** | form → API |
| `rastreio` | preço unitário que o **site mostrou** | coletor → API |

**Regra de comparação:** Δ% / Δ$ entre observação atual e a **anterior com o mesmo `sku_id` e a mesma `fonte`**.  
Misturar `compra` com `rastreio` na mesma série é **inválido**. A UI pode mostrar as duas séries lado a lado; não calcula “última vez” cruzando tipos.

Para `fonte=rastreio`, “anterior” restringe-se às observações com a **mesma `url`**. Um SKU pode ter **N URLs ativas** (multi-loja). Δ, alerta e chart **nunca** misturam URLs. Observações de URL arquivada ficam no histórico, fora do Δ/alerta/chart daquela série.

---

## 3. Superfícies da v0

### 3.0 Arquitetura de ingestão (fechada)

- **API HTTP** é a única forma de persistir preço observado (form e coletor).
- **Coletor** = processo/cron separado que, se extrair um número com sucesso, faz `POST` na API. Scraper determinístico (ou script) é o plug da v0.
- **Agente IA como runtime de preço:** fora. Não grava `preco` alucinado. IA, se existir depois, no máximo ajuda a montar seletor (D3) — nunca a série.

### 3.1 Form manual (`fonte=compra`)

- Criar/editar SKU
- Registrar preço observado **unitário** (loja, data, unidade, preço) amarrado a um SKU
- Ver histórico do SKU (série `compra`) e Δ vs última `compra`

### 3.2 Monitores multi-URL (`fonte=rastreio`)

- Mesmo produto em Araújo / Raia / Rede = **mesmo SKU** (`chave_identidade`).
- Por SKU: **N URLs ativas** permitidas (uma por loja/PDP). Mesma URL ativa duplicada = inválido.
- Job diário (worker): scrape **cada** monitor ativo → `POST` preço **unitário** `rastreio` (2x1/kit: grava unitário do frasco; total do combo fora da série).
- Se o scrape falhar: **não** inventar preço; log de falha; série intacta.
- **Arquivar URL:** remove aquela URL do Δ/alerta/chart; histórico permanece. Não arquiva as outras URLs do SKU.
- **Limiar de alerta (por monitor/URL):** `limiar_modo` = `absoluto` \| `percentual` + `limiar_valor` (mutuamente exclusivo).
  - `absoluto`: alerta se `preco <= limiar_valor` naquela URL
  - `percentual`: queda % vs última `rastreio` **da mesma URL** ≥ `limiar_valor`
  - Sem limiar naquele monitor → não alerta.
- Chart: **uma série por URL ativa** (nunca uma linha azul misturando lojas).
- Notificação: **somente Telegram**. Envio no worker.

**Contrato API para Δ/monitor multi-URL:**

- `GET /skus/{id}/observacoes/delta?fonte=rastreio` **exige** `url=` ou `monitor_id=`. Se omitido, retorna **400** com `"url or monitor_id query param required for fonte=rastreio"`. Nunca escolhe silenciosamente um monitor quando há N ativos.
- `GET /skus/{id}/monitor` sem parâmetros retorna **lista** de todos monitores ativos do SKU (pode ser 0, 1 ou N). Com `url=` ou `monitor_id=` retorna o monitor específico.
- Frontend exibe Δ **por URL** — cada monitor ativo tem seu próprio card de delta.

### 3.3 O que a v0 não promete nestas superfícies

- OCR / upload de notinha fiscal
- Catálogo multi-canal de notificação
- Matching automático SKU entre lojas (usuário amarra a `chave_identidade`)
- Ranking / “onde está mais barato agora” entre lojas do mesmo SKU
- Agente IA gravando preço
- Gravar total de kit/combo como se fosse preço unitário

---

## 4. Fora de escopo (explícito)

- OCR / parse de notinha ou cupom fiscal
- Comparação de cesta / “a compra ficou mais cara”
- WhatsApp, e-mail, push, SMS (além do Telegram único)
- Flag estruturada de promoção
- Conversão automática de unidades (kg↔g, ml↔l)
- Auth multi-usuário / times / compartilhamento (v0 = single-user)
- App mobile nativo
- Pagamentos, orçamento, metas, categorias financeiras
- Inflação agregada / índices públicos
- Matching fuzzy / ML de identidade de produto
- Agente IA como fonte de preço observado
- Scraper embutido na request web do monólito front

---

## 5. Critérios de aceite (v0 atacável)

1. Dado um SKU com ≥2 observações `compra`, Δ usa só `preco` unitário vs a `compra` imediatamente anterior (mesmo `sku_id`, mesma `fonte`).
2. `quantidade` não altera Δ nem limiar.
3. **Limiar é por monitor (URL), não por SKU.** Monitor com `limiar_modo`+valor: coleta `rastreio` **nessa URL** que satisfaz o modo dispara **uma** mensagem Telegram **referente àquela URL**; monitor sem limiar → não alerta. Nunca avalia absoluto e % ao mesmo tempo no mesmo monitor. Dois monitores no mesmo SKU podem alertar independentemente.
4. Falha de scrape não cria preço observado.
5. Comparar/alertar misturando `compra` e `rastreio`, ou misturando URLs de rastreio, é rejeitado / não oferecido.
6. Δ/alerta/chart de `rastreio` são sempre por URL; misturar URLs é rejeitado.
7. N monitores ativos no mesmo SKU são permitidos; arquivar um não afeta os outros.
8. Não existe OCR, segundo canal de notificação, ou gravação de preço por agente IA.
9. Observação de kit/combo total não entra na série unitária do SKU.
10. A v0 **não** responde “onde está mais barato agora” (mínimo / ranking entre lojas do mesmo SKU). Só perguntas **por URL** (ex.: “o Vasenol na Araújo baixou?”). Comparação cross-loja = fora de escopo até nova decisão.

---

## 6. Decisões

### D1 — Stack (fechada)

| Peça | Escolha |
|------|---------|
| API | **Go** (HTTP), domínio SKU/observação com testes nos invariantes |
| DB | **PostgreSQL** (obrigatório). Sem SQLite na v0; sem MongoDB |
| Front | **Vue 3** fino (form + histórico + limiar + URL) |
| Coletor | processo/cron **separado** (Go ou script) que só `POST` na API |
| Telegram | Bot API no **worker**, não no front |

Rejeitado na v0: Next full-stack como núcleo, scraper dentro da request web, agente IA como runtime de preço.

### Ainda abertas (sem desenhar tela/regra até fechar)

- D2 — Como o bot Telegram associa `chat_id` ao usuário único
- D3 — Estratégia de extração de preço por URL (seletor por site vs genérico)
- D4 — Retenção de histórico (ilimitado vs janela)

---

## 7. Próximo passo

1. @Socrates / @Moriaty — confirmar se os patches (unitário, limiar único, troca de URL, D1 Go) fecham a v0.
2. Após ok: implementação (API Go + Vue 3 + worker + Telegram).
