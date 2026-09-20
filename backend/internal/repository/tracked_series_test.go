package repository

import (
	"strings"
	"testing"
)

func TestListRastreioSeriesByURLSQL_PerActiveURL(t *testing.T) {
	q := listRastreioSeriesByURLSQL
	if !strings.Contains(q, "mu.ativo = TRUE") {
		t.Fatal("query must require active monitor")
	}
	if !strings.Contains(q, "mu.url = po.url") {
		t.Fatal("query must match observation url to active monitor url")
	}
	if !strings.Contains(q, "SELECT po.url") {
		t.Fatal("query must return url for per-URL series")
	}
	if !strings.Contains(q, "fonte = 'rastreio'") {
		t.Fatal("query must filter fonte=rastreio")
	}
}
