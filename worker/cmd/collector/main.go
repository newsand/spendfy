package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/newsand/spendfy/worker/internal/extractor"
	"github.com/shopspring/decimal"
)

type Config struct {
	APIURL           string
	TelegramBotToken string
	TelegramChatID   string
}

type MonitorURL struct {
	ID           int64           `json:"id"`
	SKUID        int64           `json:"sku_id"`
	URL          string          `json:"url"`
	Ativo        bool            `json:"ativo"`
	LimiarModo   *string         `json:"limiar_modo,omitempty"`
	LimiarValor  *decimal.Decimal `json:"limiar_valor,omitempty"`
	CSSSelector  *string         `json:"css_selector,omitempty"`
	RegexPattern *string         `json:"regex_pattern,omitempty"`
}

type SKU struct {
	ID            int64  `json:"id"`
	Nome          string `json:"nome"`
	UnidadePadrao string `json:"unidade_padrao"`
}

type CreateObservacaoRequest struct {
	SKUID   int64           `json:"sku_id"`
	Preco   decimal.Decimal `json:"preco"`
	Unidade string          `json:"unidade"`
	Loja    string          `json:"loja"`
	Fonte   string          `json:"fonte"`
	URL     string          `json:"url"`
}

type PrecoObservado struct {
	ID    int64           `json:"id"`
	Preco decimal.Decimal `json:"preco"`
}

type DeltaResult struct {
	CurrentPreco    decimal.Decimal  `json:"current_preco"`
	PreviousPreco   *decimal.Decimal `json:"previous_preco,omitempty"`
	DeltaAbsoluto   *decimal.Decimal `json:"delta_absoluto,omitempty"`
	DeltaPercentual *decimal.Decimal `json:"delta_percentual,omitempty"`
	HasPrevious     bool             `json:"has_previous"`
}

func main() {
	cfg := Config{
		APIURL:           getEnv("API_URL", "http://localhost:8080"),
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
	}

	if cfg.TelegramBotToken == "" || cfg.TelegramChatID == "" {
		log.Println("Warning: TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID not set, alerts disabled")
	}

	ctx := context.Background()
	collector := NewCollector(cfg)

	if err := collector.Run(ctx); err != nil {
		log.Fatalf("Collector error: %v", err)
	}
}

type Collector struct {
	cfg    Config
	client *http.Client
}

func NewCollector(cfg Config) *Collector {
	return &Collector{
		cfg: cfg,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Collector) Run(ctx context.Context) error {
	monitors, err := c.fetchActiveMonitors(ctx)
	if err != nil {
		return fmt.Errorf("fetch monitors: %w", err)
	}

	log.Printf("Found %d active monitors", len(monitors))

	for _, monitor := range monitors {
		if err := c.processMonitor(ctx, monitor); err != nil {
			log.Printf("Error processing monitor %d (SKU %d): %v", monitor.ID, monitor.SKUID, err)
		}
	}

	return nil
}

func (c *Collector) fetchActiveMonitors(ctx context.Context) ([]MonitorURL, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.cfg.APIURL+"/api/v1/monitors/active", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var monitors []MonitorURL
	if err := json.NewDecoder(resp.Body).Decode(&monitors); err != nil {
		return nil, err
	}
	return monitors, nil
}

func (c *Collector) fetchSKU(ctx context.Context, skuID int64) (*SKU, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/skus/%d", c.cfg.APIURL, skuID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sku not found: %d", skuID)
	}

	var sku SKU
	if err := json.NewDecoder(resp.Body).Decode(&sku); err != nil {
		return nil, err
	}
	return &sku, nil
}

func (c *Collector) processMonitor(ctx context.Context, monitor MonitorURL) error {
	log.Printf("Processing monitor %d: %s", monitor.ID, monitor.URL)

	sku, err := c.fetchSKU(ctx, monitor.SKUID)
	if err != nil {
		return fmt.Errorf("fetch sku: %w", err)
	}

	price, err := c.scrapePrice(ctx, monitor)
	if err != nil {
		log.Printf("Scrape failed for %s: %v", monitor.URL, err)
		return nil
	}

	log.Printf("Scraped price: %s", price.String())

	obs, err := c.postObservacao(ctx, monitor, sku, price)
	if err != nil {
		return fmt.Errorf("post observacao: %w", err)
	}

	log.Printf("Created observacao %d with price %s", obs.ID, price.String())

	if err := c.checkAndSendAlert(ctx, monitor, sku, price); err != nil {
		log.Printf("Alert check failed: %v", err)
	}

	return nil
}

func (c *Collector) scrapePrice(ctx context.Context, monitor MonitorURL) (decimal.Decimal, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", monitor.URL, nil)
	if err != nil {
		return decimal.Zero, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Spendfy/1.0)")

	resp, err := c.client.Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("fetch url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("http status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return decimal.Zero, fmt.Errorf("read body: %w", err)
	}

	ext := extractor.New()
	html := string(body)

	if monitor.CSSSelector != nil && *monitor.CSSSelector != "" {
		return ext.ExtractCSSStrict(html, *monitor.CSSSelector)
	}

	if monitor.RegexPattern != nil && *monitor.RegexPattern != "" {
		return ext.ExtractRegex(html, *monitor.RegexPattern)
	}

	return decimal.Zero, fmt.Errorf("no css_selector or regex_pattern configured for monitor %d", monitor.ID)
}

func (c *Collector) postObservacao(ctx context.Context, monitor MonitorURL, sku *SKU, price decimal.Decimal) (*PrecoObservado, error) {
	loja := extractDomain(monitor.URL)

	reqBody := CreateObservacaoRequest{
		SKUID:   monitor.SKUID,
		Preco:   price,
		Unidade: sku.UnidadePadrao,
		Loja:    loja,
		Fonte:   "rastreio",
		URL:     monitor.URL,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.cfg.APIURL+"/api/v1/observacoes", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create observacao failed: %d %s", resp.StatusCode, respBody)
	}

	var obs PrecoObservado
	if err := json.NewDecoder(resp.Body).Decode(&obs); err != nil {
		return nil, err
	}
	return &obs, nil
}

func (c *Collector) checkAndSendAlert(ctx context.Context, monitor MonitorURL, sku *SKU, currentPrice decimal.Decimal) error {
	if monitor.LimiarModo == nil || monitor.LimiarValor == nil {
		return nil
	}

	if c.cfg.TelegramBotToken == "" || c.cfg.TelegramChatID == "" {
		return nil
	}

	var shouldAlert bool
	var message string

	switch *monitor.LimiarModo {
	case "absoluto":
		if currentPrice.LessThanOrEqual(*monitor.LimiarValor) {
			shouldAlert = true
			message = fmt.Sprintf("🔔 *Alerta de Preço*\n\n*%s*\nPreço atual: R$ %s\nLimiar: R$ %s (absoluto)\n\n%s",
				sku.Nome, currentPrice.StringFixed(2), monitor.LimiarValor.StringFixed(2), monitor.URL)
		}

	case "percentual":
		delta, err := c.fetchDeltaForMonitor(ctx, monitor)
		if err != nil {
			return fmt.Errorf("fetch delta: %w", err)
		}
		if delta != nil && delta.HasPrevious && delta.DeltaPercentual != nil {
			drop := delta.DeltaPercentual.Neg()
			if drop.GreaterThanOrEqual(*monitor.LimiarValor) {
				shouldAlert = true
				message = fmt.Sprintf("🔔 *Alerta de Preço*\n\n*%s*\nPreço atual: R$ %s\nPreço anterior: R$ %s\nQueda: %.1f%%\nLimiar: %.1f%% (percentual)\n\n%s",
					sku.Nome, currentPrice.StringFixed(2), delta.PreviousPreco.StringFixed(2),
					drop.InexactFloat64(), monitor.LimiarValor.InexactFloat64(), monitor.URL)
			}
		}
	}

	if shouldAlert {
		log.Printf("Sending alert for SKU %d: %s", sku.ID, sku.Nome)
		return c.sendTelegramMessage(ctx, message)
	}

	return nil
}

func (c *Collector) fetchDeltaForMonitor(ctx context.Context, monitor MonitorURL) (*DeltaResult, error) {
	deltaURL := fmt.Sprintf("%s/api/v1/skus/%d/observacoes/delta?fonte=rastreio&url=%s",
		c.cfg.APIURL, monitor.SKUID, url.QueryEscape(monitor.URL))
	req, err := http.NewRequestWithContext(ctx, "GET", deltaURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode == http.StatusBadRequest {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("delta status %d", resp.StatusCode)
	}

	var delta DeltaResult
	if err := json.NewDecoder(resp.Body).Decode(&delta); err != nil {
		return nil, err
	}
	return &delta, nil
}

func (c *Collector) sendTelegramMessage(ctx context.Context, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", c.cfg.TelegramBotToken)

	data := url.Values{}
	data.Set("chat_id", c.cfg.TelegramChatID)
	data.Set("text", message)
	data.Set("parse_mode", "Markdown")

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram error: %d %s", resp.StatusCode, body)
	}

	log.Println("Telegram message sent successfully")
	return nil
}

func extractDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Host
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
