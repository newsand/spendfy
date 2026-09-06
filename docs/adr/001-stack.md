# ADR-001: Stack PostgreSQL + Go + Vue 3

## Status

Accepted (D1 fechada)

## Context

Spendfy v0 precisa de uma stack simples para single-user, com separação clara entre API, front e worker.

## Decision

| Componente | Escolha | Motivo |
|------------|---------|--------|
| API | Go 1.22 | Performance, tipagem, deploy simples (binário único) |
| DB | PostgreSQL 16 | Robusto, ACID, suporte a NUMERIC para preços, escalável se necessário |
| Front | Vue 3 + Vite | Leve, reativo, dev experience moderna |
| Worker | Go (separado) | Reutiliza tipos, sem dependência do ciclo HTTP |

## Rejected

- **SQLite**: Originalmente considerado para v0, mas PostgreSQL oferece melhor suporte a concorrência caso o worker rode em paralelo com a API, e facilita migração futura.
- **Next.js full-stack**: Complexidade desnecessária para v0. Scraper dentro de request web viola a arquitetura definida.
- **MongoDB**: Não há necessidade de schema flexível; preços são estruturados.

## Consequences

- Requer PostgreSQL rodando (docker-compose simplifica)
- Deploy em produção precisa de container ou VM com Go e Postgres
- Worker é processo separado, pode ser cron ou serviço
