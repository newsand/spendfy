-- Spendfy v0 Schema for PostgreSQL

CREATE TABLE IF NOT EXISTS skus (
    id BIGSERIAL PRIMARY KEY,
    nome TEXT NOT NULL,
    marca TEXT,
    tamanho_ou_variante TEXT,
    unidade_padrao TEXT NOT NULL CHECK (unidade_padrao IN ('un', 'kg', 'g', 'l', 'ml')),
    chave_identidade TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS precos_observados (
    id BIGSERIAL PRIMARY KEY,
    sku_id BIGINT NOT NULL REFERENCES skus(id) ON DELETE CASCADE,
    preco NUMERIC(15, 4) NOT NULL CHECK (preco > 0),
    moeda TEXT NOT NULL DEFAULT 'BRL',
    quantidade NUMERIC(15, 4),
    unidade TEXT NOT NULL CHECK (unidade IN ('un', 'kg', 'g', 'l', 'ml')),
    loja TEXT NOT NULL,
    data TIMESTAMPTZ NOT NULL,
    fonte TEXT NOT NULL CHECK (fonte IN ('compra', 'rastreio')),
    url TEXT,
    notas TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT rastreio_requires_url CHECK (
        fonte != 'rastreio' OR url IS NOT NULL
    )
);

CREATE INDEX idx_precos_observados_sku_fonte ON precos_observados(sku_id, fonte, data DESC);
CREATE INDEX idx_precos_observados_sku_fonte_url ON precos_observados(sku_id, fonte, url, data DESC);

CREATE TABLE IF NOT EXISTS monitor_urls (
    id BIGSERIAL PRIMARY KEY,
    sku_id BIGINT NOT NULL REFERENCES skus(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    ativo BOOLEAN NOT NULL DEFAULT TRUE,
    limiar_modo TEXT CHECK (limiar_modo IS NULL OR limiar_modo IN ('absoluto', 'percentual')),
    limiar_valor NUMERIC(15, 4),
    css_selector TEXT,
    regex_pattern TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    archived_at TIMESTAMPTZ,
    CONSTRAINT limiar_both_or_none CHECK (
        (limiar_modo IS NULL AND limiar_valor IS NULL) OR
        (limiar_modo IS NOT NULL AND limiar_valor IS NOT NULL)
    )
);

CREATE UNIQUE INDEX idx_monitor_urls_sku_ativo ON monitor_urls(sku_id) WHERE ativo = TRUE;

CREATE TABLE IF NOT EXISTS coleta_logs (
    id BIGSERIAL PRIMARY KEY,
    monitor_id BIGINT NOT NULL REFERENCES monitor_urls(id) ON DELETE CASCADE,
    success BOOLEAN NOT NULL,
    preco_observado_id BIGINT REFERENCES precos_observados(id),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_coleta_logs_monitor ON coleta_logs(monitor_id, created_at DESC);
