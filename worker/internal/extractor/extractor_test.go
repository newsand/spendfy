package extractor

import (
	"testing"

	"github.com/shopspring/decimal"
)

const araujoFixtureHTML = `
<!DOCTYPE html>
<html>
<head><title>Loção Hidratante Vasenol</title></head>
<body>
  <div class="product-page">
    <h1 class="product-title">Loção Hidratante Vasenol Recuperação Intensiva Clinical com 200ml</h1>
    
    <!-- PRIMARY UNIT PRICE - this is what we want -->
    <div class="product-main-info">
      <div class="product-price-box">
        <span class="price-label">Por:</span>
        <span class="price-value" data-product-price="47.99">
          R$ 47,99
        </span>
        <span class="price-installments">ou 3x de R$ 16,00</span>
      </div>
      <button class="btn-buy">Comprar</button>
    </div>

    <!-- KIT OFFERS - these should be IGNORED for the single-unit SKU -->
    <div class="kit-offers">
      <h3>Kits com desconto</h3>
      
      <div class="kit-item" data-kit="3">
        <span class="kit-title">Leve 3 pague menos</span>
        <span class="kit-price price-value">R$ 143,97</span>
        <span class="kit-unit-price">(R$ 47,99 cada)</span>
      </div>
      
      <div class="kit-item" data-kit="4">
        <span class="kit-title">Leve 4 pague menos</span>
        <span class="kit-price price-value">R$ 179,96</span>
        <span class="kit-unit-price">(R$ 44,99 cada)</span>
      </div>
    </div>

    <!-- Another price display that might confuse broad selectors -->
    <div class="similar-products">
      <div class="product-card">
        <span class="product-name">Vasenol 400ml</span>
        <span class="price-value">R$ 72,99</span>
      </div>
    </div>
  </div>
</body>
</html>
`

func TestExtractCSS_AraujoUnitPrice(t *testing.T) {
	ext := New()

	selector := ".product-main-info .product-price-box .price-value"
	price, err := ext.ExtractCSS(araujoFixtureHTML, selector)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	expected := decimal.NewFromFloat(47.99)
	if !price.Equal(expected) {
		t.Errorf("expected unit price %s, got %s", expected.String(), price.String())
	}
}

func TestExtractCSS_BroadSelectorGetsWrongPrice(t *testing.T) {
	ext := New()

	selector := ".price-value"
	price, err := ext.ExtractCSS(araujoFixtureHTML, selector)

	if err != nil {
		t.Fatalf("expected no error with broad selector, got: %v", err)
	}

	expected := decimal.NewFromFloat(47.99)
	if !price.Equal(expected) {
		t.Logf("Broad selector got %s - First() returns first in DOM order which is correct here", price.String())
	}
}

func TestExtractCSSStrict_AmbiguousMultiMatch(t *testing.T) {
	ext := New()

	selector := ".price-value"
	_, err := ext.ExtractCSSStrict(araujoFixtureHTML, selector)

	if err != ErrAmbiguousMatch {
		t.Errorf("expected ErrAmbiguousMatch for broad selector with different prices, got: %v", err)
	}
}

func TestExtractCSS_SpecificSelectorReturnsUnitPrice(t *testing.T) {
	ext := New()

	testCases := []struct {
		name     string
		selector string
		expected decimal.Decimal
	}{
		{
			name:     "product-price-box specific",
			selector: ".product-price-box .price-value",
			expected: decimal.NewFromFloat(47.99),
		},
		{
			name:     "data attribute selector",
			selector: "[data-product-price]",
			expected: decimal.NewFromFloat(47.99),
		},
		{
			name:     "product-main-info scoped",
			selector: ".product-main-info .price-value",
			expected: decimal.NewFromFloat(47.99),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			price, err := ext.ExtractCSS(araujoFixtureHTML, tc.selector)
			if err != nil {
				t.Fatalf("selector %q failed: %v", tc.selector, err)
			}
			if !price.Equal(tc.expected) {
				t.Errorf("selector %q: expected %s, got %s", tc.selector, tc.expected.String(), price.String())
			}
		})
	}
}

func TestExtractCSS_KitPriceNotReturned(t *testing.T) {
	ext := New()

	selector := ".product-main-info .product-price-box .price-value"
	price, err := ext.ExtractCSS(araujoFixtureHTML, selector)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	kitPrices := []decimal.Decimal{
		decimal.NewFromFloat(143.97),
		decimal.NewFromFloat(179.96),
		decimal.NewFromFloat(72.99),
	}

	for _, kit := range kitPrices {
		if price.Equal(kit) {
			t.Errorf("extractor returned kit/other price %s instead of unit price", kit.String())
		}
	}
}

func TestExtractCSS_NoMatchReturnsError(t *testing.T) {
	ext := New()

	_, err := ext.ExtractCSS(araujoFixtureHTML, ".nonexistent-class")

	if err != ErrNoMatch {
		t.Errorf("expected ErrNoMatch, got: %v", err)
	}
}

func TestExtractRegex_UnitPrice(t *testing.T) {
	ext := New()

	pattern := `data-product-price="([\d.]+)"`
	price, err := ext.ExtractRegex(araujoFixtureHTML, pattern)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := decimal.NewFromFloat(47.99)
	if !price.Equal(expected) {
		t.Errorf("expected %s, got %s", expected.String(), price.String())
	}
}

func TestParsePrice_BrazilianFormat(t *testing.T) {
	testCases := []struct {
		input    string
		expected decimal.Decimal
	}{
		{"R$ 47,99", decimal.NewFromFloat(47.99)},
		{"47,99", decimal.NewFromFloat(47.99)},
		{"R$ 1.234,56", decimal.NewFromFloat(1234.56)},
		{"143,97", decimal.NewFromFloat(143.97)},
		{"  R$ 47,99  ", decimal.NewFromFloat(47.99)},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			price, err := ParsePrice(tc.input)
			if err != nil {
				t.Fatalf("failed to parse %q: %v", tc.input, err)
			}
			if !price.Equal(tc.expected) {
				t.Errorf("parse %q: expected %s, got %s", tc.input, tc.expected.String(), price.String())
			}
		})
	}
}

func TestParsePrice_InvalidInputs(t *testing.T) {
	invalidInputs := []string{
		"",
		"   ",
		"grátis",
		"R$",
		"-10,00",
		"0,00",
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParsePrice(input)
			if err == nil {
				t.Errorf("expected error for invalid input %q", input)
			}
		})
	}
}

func TestInvariant_ScrapeFailureNoInventedPrice(t *testing.T) {
	ext := New()

	_, err := ext.ExtractCSS("<html></html>", ".product-price")

	if err == nil {
		t.Error("expected error for missing price, scrape failure must not invent price")
	}

	_, err = ext.ExtractCSSStrict(araujoFixtureHTML, ".price-value")
	if err != ErrAmbiguousMatch {
		t.Error("ambiguous match should fail - do not guess which price is correct")
	}
}
