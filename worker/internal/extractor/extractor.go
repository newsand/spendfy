package extractor

import (
	"errors"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/shopspring/decimal"
)

var (
	ErrNoMatch        = errors.New("selector matched nothing")
	ErrAmbiguousMatch = errors.New("selector matched multiple prices - ambiguous")
	ErrInvalidRegex   = errors.New("invalid regex pattern")
	ErrParsePrice     = errors.New("could not parse price from text")
)

type Extractor struct{}

func New() *Extractor {
	return &Extractor{}
}

func (e *Extractor) ExtractCSS(html, selector string) (decimal.Decimal, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return decimal.Zero, err
	}

	selection := doc.Find(selector)
	if selection.Length() == 0 {
		return decimal.Zero, ErrNoMatch
	}

	text := selection.First().Text()
	if text == "" {
		return decimal.Zero, ErrNoMatch
	}

	return ParsePrice(text)
}

func (e *Extractor) ExtractCSSStrict(html, selector string) (decimal.Decimal, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return decimal.Zero, err
	}

	selection := doc.Find(selector)
	count := selection.Length()

	if count == 0 {
		return decimal.Zero, ErrNoMatch
	}

	if count > 1 {
		prices := make([]decimal.Decimal, 0, count)
		selection.Each(func(i int, s *goquery.Selection) {
			if p, err := ParsePrice(s.Text()); err == nil && p.GreaterThan(decimal.Zero) {
				prices = append(prices, p)
			}
		})

		if len(prices) > 1 {
			first := prices[0]
			allSame := true
			for _, p := range prices[1:] {
				if !p.Equal(first) {
					allSame = false
					break
				}
			}
			if !allSame {
				return decimal.Zero, ErrAmbiguousMatch
			}
		}
	}

	text := selection.First().Text()
	return ParsePrice(text)
}

func (e *Extractor) ExtractRegex(html, pattern string) (decimal.Decimal, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return decimal.Zero, ErrInvalidRegex
	}

	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return decimal.Zero, ErrNoMatch
	}

	return ParsePrice(matches[1])
}

func (e *Extractor) ExtractDefault(html string) (decimal.Decimal, error) {
	patterns := []string{
		`"price":\s*"?([\d.,]+)"?`,
		`data-price="([\d.,]+)"`,
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p)
		matches := re.FindStringSubmatch(html)
		if len(matches) >= 2 {
			price, err := ParsePrice(matches[1])
			if err == nil && price.GreaterThan(decimal.Zero) {
				return price, nil
			}
		}
	}

	return decimal.Zero, ErrNoMatch
}

func ParsePrice(s string) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "R$", "")
	s = strings.ReplaceAll(s, "\u00a0", "")

	if s == "" {
		return decimal.Zero, ErrParsePrice
	}

	hasDot := strings.Contains(s, ".")
	hasComma := strings.Contains(s, ",")

	if hasDot && hasComma {
		if strings.LastIndex(s, ",") > strings.LastIndex(s, ".") {
			s = strings.ReplaceAll(s, ".", "")
			s = strings.ReplaceAll(s, ",", ".")
		} else {
			s = strings.ReplaceAll(s, ",", "")
		}
	} else if hasComma {
		s = strings.ReplaceAll(s, ",", ".")
	}

	price, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, ErrParsePrice
	}

	if price.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, ErrParsePrice
	}

	return price, nil
}
