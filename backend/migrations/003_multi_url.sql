-- Multi-URL: allow N active monitors per SKU (unique per sku+url when active)
DROP INDEX IF EXISTS idx_monitor_urls_sku_ativo;
CREATE UNIQUE INDEX idx_monitor_urls_sku_url_ativo
  ON monitor_urls(sku_id, url) WHERE ativo = TRUE;
