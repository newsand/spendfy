-- Seed data for testing
-- Run after 001_init.sql

INSERT INTO skus (nome, marca, tamanho_ou_variante, unidade_padrao, chave_identidade)
VALUES (
    'Loção hidratante Vasenol Clinical 200ml',
    'Vasenol',
    '200ml',
    'un',
    'vasenol-clinical-200ml'
) ON CONFLICT (chave_identidade) DO NOTHING;

-- Set up monitor URL for the seed SKU
-- IMPORTANT: Selector must target PRIMARY UNIT PRICE only, not kit/combo prices.
-- The Araújo PDP shows multiple prices (unit, kit-3, kit-4). We want only the main buy price.
-- Selector: .product-main-info .product-price-box .price-value
-- Alternative: [data-product-price] if available
--
-- Limiar: absoluto R$ 45.00 — typical shelf price is ~R$ 47.99, so alert fires
-- when price drops to R$ 45 or below (a ~6% discount).
INSERT INTO monitor_urls (sku_id, url, ativo, css_selector, limiar_modo, limiar_valor)
SELECT 
    id,
    'https://www.araujo.com.br/locao-hidratante-vasenol-recuperacao-intensiva-clinical-com-200ml/73586.html',
    TRUE,
    '.product-main-info .product-price-box .price-value',
    'absoluto',
    45.00
FROM skus 
WHERE chave_identidade = 'vasenol-clinical-200ml'
ON CONFLICT DO NOTHING;
