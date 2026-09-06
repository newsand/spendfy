package repository

// listRastreioSeriesForChartSQL returns rastreio points only for the active monitor URL.
// Archived monitor URLs must not appear on the dashboard chart (spec parity with Δ/alerta).
const listRastreioSeriesForChartSQL = `
			SELECT po.data, po.preco FROM precos_observados po
			INNER JOIN monitor_urls mu ON mu.sku_id = po.sku_id AND mu.ativo = TRUE AND mu.url = po.url
			WHERE po.sku_id = $1 AND po.fonte = 'rastreio'
			ORDER BY po.data ASC
		`
