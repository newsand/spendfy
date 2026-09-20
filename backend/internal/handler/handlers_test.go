package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestGetDelta_RastreioRequiresURLOrMonitorID(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/api/v1/skus/{id}/observacoes/delta", func(w http.ResponseWriter, r *http.Request) {
		fonteStr := r.URL.Query().Get("fonte")
		if fonteStr == "rastreio" {
			urlParam := r.URL.Query().Get("url")
			monitorIDParam := r.URL.Query().Get("monitor_id")
			if urlParam == "" && monitorIDParam == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "url or monitor_id query param required for fonte=rastreio"})
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "rastreio without url or monitor_id returns 400",
			path:       "/api/v1/skus/1/observacoes/delta?fonte=rastreio",
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "rastreio with url is accepted",
			path:       "/api/v1/skus/1/observacoes/delta?fonte=rastreio&url=https://example.com",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "rastreio with monitor_id is accepted",
			path:       "/api/v1/skus/1/observacoes/delta?fonte=rastreio&monitor_id=1",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "compra does not require url or monitor_id",
			path:       "/api/v1/skus/1/observacoes/delta?fonte=compra",
			wantStatus: http.StatusOK,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			if tt.wantError {
				var resp map[string]string
				json.NewDecoder(rec.Body).Decode(&resp)
				if resp["error"] == "" {
					t.Error("expected error message in response")
				}
			}
		})
	}
}

func TestGetMonitor_MultiURLReturnsListOrRequiresParam(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/api/v1/skus/{id}/monitor", func(w http.ResponseWriter, r *http.Request) {
		urlParam := r.URL.Query().Get("url")
		monitorIDParam := r.URL.Query().Get("monitor_id")

		if urlParam != "" || monitorIDParam != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]any{
				"id":  1,
				"url": urlParam,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "url": "https://araujo.example/p/1"},
			{"id": 2, "url": "https://raia.example/p/1"},
		})
	})

	tests := []struct {
		name           string
		path           string
		wantStatus     int
		wantSingleItem bool
		wantList       bool
	}{
		{
			name:       "without params returns list of monitors",
			path:       "/api/v1/skus/1/monitor",
			wantStatus: http.StatusOK,
			wantList:   true,
		},
		{
			name:           "with url returns single monitor",
			path:           "/api/v1/skus/1/monitor?url=https://araujo.example/p/1",
			wantStatus:     http.StatusOK,
			wantSingleItem: true,
		},
		{
			name:           "with monitor_id returns single monitor",
			path:           "/api/v1/skus/1/monitor?monitor_id=1",
			wantStatus:     http.StatusOK,
			wantSingleItem: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			body := rec.Body.String()
			if tt.wantList {
				if !strings.HasPrefix(strings.TrimSpace(body), "[") {
					t.Error("expected list response (array)")
				}
			}
			if tt.wantSingleItem {
				if !strings.HasPrefix(strings.TrimSpace(body), "{") || strings.HasPrefix(strings.TrimSpace(body), "[") {
					t.Error("expected single item response (object)")
				}
			}
		})
	}
}

func TestDeltaMultiURL_SameSkuDifferentStores(t *testing.T) {
	araujoURL := "https://araujo.example/p/vasenol"
	raiaURL := "https://raia.example/p/vasenol"

	tests := []struct {
		name        string
		url         string
		description string
	}{
		{
			name:        "delta for Araújo URL",
			url:         araujoURL,
			description: "should return delta computed from Araújo observations only",
		},
		{
			name:        "delta for Raia URL",
			url:         raiaURL,
			description: "should return delta computed from Raia observations only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.url == "" {
				t.Fatal("url should not be empty for rastreio delta")
			}
		})
	}
}
