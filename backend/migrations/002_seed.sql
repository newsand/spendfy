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
INSERT INTO monitor_urls (sku_id, url, ativo, css_selector)
SELECT 
    id,
    'https://www.araujo.com.br/locao-hidratante-vasenol-recuperacao-intensiva-clinical-com-200ml/73586.html',
    TRUE,
    '.product-main-info .product-price-box .price-value'
FROM skus 
WHERE chave_identidade = 'vasenol-clinical-200ml'
ON CONFLICT DO NOTHING;
