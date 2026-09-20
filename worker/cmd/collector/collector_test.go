package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
)

const kitFirstHTML = `
<!DOCTYPE html>
<html>
<body>
  <!-- KIT PRICE APPEARS FIRST IN DOM - this is the trap -->
  <div class="kit-offer">
    <span class="price-value">R$ 143,97</span>
    <span>Kit com 3 unidades</span>
  </div>
  
  <!-- UNIT PRICE APPEARS SECOND -->
  <div class="product-main">
    <span class="price-value">R$ 47,99</span>
    <button>Comprar</button>
  </div>
</body>
</html>
`

const unitOnlyHTML = `
<!DOCTYPE html>
<html>
<body>
  <div class="product-main-info">
    <div class="product-price-box">
      <span class="price-value">R$ 47,99</span>
    </div>
  </div>
</body>
</html>
`

func TestCollector_ScrapePriceRejectsBroadSelectorWithKitFirst(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(kitFirstHTML))
	}))
	defer server.Close()

	collector := NewCollector(Config{APIURL: "http://localhost:8080"})

	broadSelector := ".price-value"
	monitor := MonitorURL{
		ID:          1,
		SKUID:       1,
		URL:         server.URL,
		CSSSelector: &broadSelector,
	}

	_, err := collector.scrapePrice(context.Background(), monitor)

	if err == nil {
		t.Fatal("expected error for broad selector with kit-first DOM, got nil - would have posted kit price 143.97")
	}

	t.Logf("correctly rejected: %v", err)
}

func TestCollector_ScrapePriceAcceptsSpecificSelector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(unitOnlyHTML))
	}))
	defer server.Close()

	collector := NewCollector(Config{APIURL: "http://localhost:8080"})

	specificSelector := ".product-main-info .product-price-box .price-value"
	monitor := MonitorURL{
		ID:          1,
		SKUID:       1,
		URL:         server.URL,
		CSSSelector: &specificSelector,
	}

	price, err := collector.scrapePrice(context.Background(), monitor)

	if err != nil {
		t.Fatalf("expected no error for specific selector, got: %v", err)
	}

	expected := decimal.NewFromFloat(47.99)
	if !price.Equal(expected) {
		t.Errorf("expected price %s, got %s", expected.String(), price.String())
	}
}

func TestCollector_ScrapePriceRejectsNoSelector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(unitOnlyHTML))
	}))
	defer server.Close()

	collector := NewCollector(Config{APIURL: "http://localhost:8080"})

	monitor := MonitorURL{
		ID:          1,
		SKUID:       1,
		URL:         server.URL,
		CSSSelector: nil,
	}

	_, err := collector.scrapePrice(context.Background(), monitor)

	if err == nil {
		t.Fatal("expected error when no selector configured, got nil")
	}

	t.Logf("correctly rejected no-selector: %v", err)
}

func TestCollector_ScrapePriceWithRegexWorks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><span data-unit-price="47.99">Price</span></body></html>`))
	}))
	defer server.Close()

	collector := NewCollector(Config{APIURL: "http://localhost:8080"})

	regex := `data-unit-price="([\d.]+)"`
	monitor := MonitorURL{
		ID:           1,
		SKUID:        1,
		URL:          server.URL,
		RegexPattern: &regex,
	}

	price, err := collector.scrapePrice(context.Background(), monitor)

	if err != nil {
		t.Fatalf("expected no error for regex, got: %v", err)
	}

	expected := decimal.NewFromFloat(47.99)
	if !price.Equal(expected) {
		t.Errorf("expected price %s, got %s", expected.String(), price.String())
	}
}

func TestCollector_FetchDeltaForMonitorIncludesURL(t *testing.T) {
	var capturedURL string
	var capturedQuery string

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.Path
		capturedQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusNotFound)
	}))
	defer apiServer.Close()

	collector := NewCollector(Config{APIURL: apiServer.URL})

	monitor := MonitorURL{
		ID:    42,
		SKUID: 123,
		URL:   "https://araujo.example/p/vasenol",
	}

	_, _ = collector.fetchDeltaForMonitor(context.Background(), monitor)

	if capturedURL != "/api/v1/skus/123/observacoes/delta" {
		t.Errorf("expected delta path for SKU 123, got %s", capturedURL)
	}

	if capturedQuery == "" {
		t.Fatal("expected query params, got empty")
	}

	if !contains(capturedQuery, "fonte=rastreio") {
		t.Errorf("expected fonte=rastreio in query, got %s", capturedQuery)
	}

	if !contains(capturedQuery, "url=") {
		t.Errorf("expected url= param in query (per-monitor delta), got %s", capturedQuery)
	}

	if !contains(capturedQuery, "araujo.example") {
		t.Errorf("expected monitor URL in query, got %s", capturedQuery)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
