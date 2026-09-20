package repository

// listRastreioSeriesByURLSQL returns rastreio points for each ACTIVE monitor URL on the SKU.
// One series per URL — never merge Araújo + Raia into a single line.
const listRastreioSeriesByURLSQL = `
			SELECT po.url, po.data, po.preco FROM precos_observados po
			INNER JOIN monitor_urls mu ON mu.sku_id = po.sku_id AND mu.ativo = TRUE AND mu.url = po.url
			WHERE po.sku_id = $1 AND po.fonte = 'rastreio'
			ORDER BY po.url, po.data ASC
		`

// Deprecated alias kept for older tests naming.
const listRastreioSeriesForChartSQL = listRastreioSeriesByURLSQL
