package repository

import (
	"strings"
	"testing"
)

func TestListRastreioSeriesForChartSQL_FiltersActiveURL(t *testing.T) {
	q := listRastreioSeriesForChartSQL
	if !strings.Contains(q, "mu.ativo = TRUE") {
		t.Fatal("query must require active monitor")
	}
	if !strings.Contains(q, "mu.url = po.url") {
		t.Fatal("query must match observation url to active monitor url")
	}
	if !strings.Contains(q, "fonte = 'rastreio'") {
		t.Fatal("query must filter fonte=rastreio")
	}
	// Must NOT be the unfiltered form
	bad := "WHERE sku_id = $1 AND fonte = 'rastreio'"
	if strings.Contains(q, bad) && !strings.Contains(q, "INNER JOIN monitor_urls") {
		t.Fatal("unfiltered rastreio query")
	}
}
