# ADR-002: Telegram single-user via env var

## Status

Accepted (D2 fechada para v0)

## Context

v0 é single-user. Precisamos de um mecanismo simples para associar o único usuário ao chat do Telegram para receber alertas.

## Decision

Usar variáveis de ambiente:

- `TELEGRAM_BOT_TOKEN` — Token do bot criado via @BotFather
- `TELEGRAM_CHAT_ID` — ID do chat do usuário único

O worker lê essas variáveis e envia mensagens diretamente via Bot API quando o limiar de alerta é atingido.

## Rationale

- Simplicidade: não requer banco adicional para associação user↔chat
- Segurança: tokens não ficam no código
- Single-user: não há necessidade de mapeamento dinâmico

## Consequences

- Multi-user futuro precisará de nova solução (tabela de notificações, OAuth, etc.)
- Usuário precisa configurar manualmente o chat_id (documentado no README)
- Se as variáveis não estiverem configuradas, alertas são silenciosamente desabilitados

## How to get chat_id

1. Criar bot com @BotFather
2. Enviar qualquer mensagem para o bot
3. Acessar `https://api.telegram.org/bot<TOKEN>/getUpdates`
4. Copiar o `chat.id` do JSON retornado
