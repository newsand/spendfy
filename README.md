# Spendfy

App pessoal para acompanhar preço de **SKU** ao longo do tempo (compra registrada + rastreio de 1 URL).

## Spec

Ver [SPENDFY-ESPEC-v0.md](./SPENDFY-ESPEC-v0.md).

## Stack (D1 fechada)

| Componente | Tecnologia |
|------------|------------|
| API | Go 1.22 (HTTP, chi router) |
| Banco | PostgreSQL 16 |
| Front | Vue 3 + Vite |
| Worker | Go (processo separado, cron) |
| Notificações | Telegram Bot API |

## Arquitetura

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Vue 3     │────▶│   Go API    │────▶│  PostgreSQL │
│  Frontend   │     │   :8080     │     │   :5432     │
└─────────────┘     └─────────────┘     └─────────────┘
                           ▲
                           │ POST /api/v1/observacoes
                           │
                    ┌──────┴──────┐
                    │   Worker    │──────▶ Telegram
                    │  (cron)     │
                    └─────────────┘
```

- **API HTTP** é a única forma de persistir preço observado
- **Worker** = processo separado que faz scrape e POST na API
- **Telegram** enviado pelo worker, não pelo front

## Quick Start (Docker Compose)

```bash
# Subir todos os serviços
docker-compose up -d

# API: http://localhost:8080
# Frontend: http://localhost:5173
```

## Dashboard de Produtos Rastreados

A tela inicial do frontend (`/`) exibe o **Dashboard de Produtos Rastreados**:

- Lista SKUs que possuem observações de preço ou monitor ativo
- Cada card mostra:
  - Nome, marca e variante do SKU
  - Badge de "Monitor ativo" se houver URL configurada
  - **Último valor pago** (fonte=compra) ou "—" se não houver compra registrada
  - **Gráfico de preço ao longo do tempo** com séries separadas para compra (verde) e rastreio (azul)
- Clique no card para ver detalhes do SKU

A navegação superior permite alternar entre Dashboard e lista de SKUs.

## Desenvolvimento Local

### Pré-requisitos

- Go 1.22+
- Node.js 22+
- PostgreSQL 16+

### Banco de Dados

```bash
# Criar banco
createdb spendfy

# Rodar migrations
psql spendfy < backend/migrations/001_init.sql
```

Ou via Docker:

```bash
docker-compose up -d postgres
```

### API

```bash
cd backend
go mod download
DATABASE_URL="postgres://spendfy:spendfy@localhost:5432/spendfy?sslmode=disable" go run ./cmd/api
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Worker/Collector

```bash
cd worker
go mod download
API_URL="http://localhost:8080" go run ./cmd/collector
```

## Variáveis de Ambiente

### API

| Variável | Descrição | Default |
|----------|-----------|---------|
| `DATABASE_URL` | Connection string PostgreSQL | `postgres://spendfy:spendfy@localhost:5432/spendfy?sslmode=disable` |
| `PORT` | Porta HTTP | `8080` |

### Worker

| Variável | Descrição | Default |
|----------|-----------|---------|
| `API_URL` | URL da API | `http://localhost:8080` |
| `TELEGRAM_BOT_TOKEN` | Token do bot Telegram | (obrigatório para alertas) |
| `TELEGRAM_CHAT_ID` | Chat ID para enviar alertas | (obrigatório para alertas) |

## Testes

```bash
cd backend
go test ./...
```

## API Endpoints

### SKUs

- `GET /api/v1/skus` — Listar SKUs
- `POST /api/v1/skus` — Criar SKU
- `GET /api/v1/skus/:id` — Obter SKU
- `PUT /api/v1/skus/:id` — Atualizar SKU
- `DELETE /api/v1/skus/:id` — Excluir SKU

### Observações

- `POST /api/v1/observacoes` — Criar observação (compra ou rastreio)
- `GET /api/v1/observacoes/:id` — Obter observação
- `GET /api/v1/skus/:id/observacoes` — Listar observações do SKU (query: `?fonte=compra|rastreio`)
- `GET /api/v1/skus/:id/observacoes/delta` — Obter delta vs última observação (query: `?fonte=compra|rastreio`)

### Monitor (rastreio)

- `GET /api/v1/skus/:id/monitor` — Obter monitor ativo
- `POST /api/v1/skus/:id/monitor` — Configurar/trocar URL do monitor (arquiva anterior)
- `PUT /api/v1/skus/:id/monitor/limiar` — Atualizar limiar de alerta
- `GET /api/v1/skus/:id/monitor/history` — Histórico de URLs
- `GET /api/v1/monitors/active` — Listar todos os monitors ativos (usado pelo worker)

### Dashboard

- `GET /api/v1/dashboard/tracked-skus` — Lista SKUs com observações ou monitor ativo, incluindo último preço de compra e séries de preço (compra e rastreio separadas)

## Configuração do Telegram (D2)

Para v0 single-user, o `chat_id` é configurado via variável de ambiente.

1. Crie um bot com [@BotFather](https://t.me/botfather)
2. Obtenha o token
3. Envie uma mensagem para o bot
4. Obtenha seu `chat_id` via `https://api.telegram.org/bot<TOKEN>/getUpdates`
5. Configure as variáveis:

```bash
export TELEGRAM_BOT_TOKEN="123456:ABC..."
export TELEGRAM_CHAT_ID="987654321"
```

## Extração de Preço (D3)

O worker suporta três modos de extração por monitor:

1. **CSS Selector** — Ex: `.price`, `#product-price`
2. **Regex Pattern** — Ex: `R\$\s*([\d.,]+)` (primeiro grupo capturado)
3. **Default** — Tenta padrões comuns (R$, data-price, class="price")

Configure ao criar/trocar o monitor via API ou frontend.

## Decisões de Arquitetura

Ver [docs/adr/](./docs/adr/) para ADRs completos.

- [ADR-001: Stack PostgreSQL + Go + Vue 3](./docs/adr/001-stack.md)
- [ADR-002: Telegram single-user via env var](./docs/adr/002-telegram.md)
- [ADR-003: Extração de preço por seletor](./docs/adr/003-extractor.md)
