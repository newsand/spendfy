package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/newsand/spendfy/backend/internal/domain"
	"github.com/newsand/spendfy/backend/internal/service"
)

type Handlers struct {
	svc *service.Service
}

func New(svc *service.Service) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/skus", func(r chi.Router) {
			r.Get("/", h.ListSKUs)
			r.Post("/", h.CreateSKU)
			r.Get("/{id}", h.GetSKU)
			r.Put("/{id}", h.UpdateSKU)
			r.Delete("/{id}", h.DeleteSKU)

			r.Get("/{id}/observacoes", h.ListObservacoes)
			r.Get("/{id}/observacoes/delta", h.GetDelta)

			r.Get("/{id}/monitor", h.GetMonitor)
			r.Post("/{id}/monitor", h.SetMonitor)
			r.Put("/{id}/monitor/limiar", h.UpdateMonitorLimiar)
			r.Get("/{id}/monitor/history", h.ListMonitorHistory)
		})

		r.Route("/observacoes", func(r chi.Router) {
			r.Post("/", h.CreateObservacao)
			r.Get("/{id}", h.GetObservacao)
		})

		r.Get("/monitors/active", h.ListActiveMonitors)
	})
}

func (h *Handlers) ListSKUs(w http.ResponseWriter, r *http.Request) {
	skus, err := h.svc.ListSKUs(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, skus)
}

func (h *Handlers) CreateSKU(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateSKURequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	sku, err := h.svc.CreateSKU(r.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrChaveIdentidadeExists) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		var ve *domain.ValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sku)
}

func (h *Handlers) GetSKU(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	sku, err := h.svc.GetSKU(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrSKUNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sku)
}

func (h *Handlers) UpdateSKU(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req domain.UpdateSKURequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	sku, err := h.svc.UpdateSKU(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, domain.ErrSKUNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sku)
}

func (h *Handlers) DeleteSKU(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.DeleteSKU(r.Context(), id); err != nil {
		if errors.Is(err, domain.ErrSKUNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) CreateObservacao(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateObservacaoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	obs, err := h.svc.CreateObservacao(r.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrSKUNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		if errors.Is(err, domain.ErrUnidadeMismatch) || errors.Is(err, domain.ErrURLRequiredForRastreio) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		var ve *domain.ValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, obs)
}

func (h *Handlers) GetObservacao(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	obs, err := h.svc.GetObservacao(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrObservacaoNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, obs)
}

func (h *Handlers) ListObservacoes(w http.ResponseWriter, r *http.Request) {
	skuID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var fonte *domain.Fonte
	if f := r.URL.Query().Get("fonte"); f != "" {
		ft := domain.Fonte(f)
		if !ft.Valid() {
			writeError(w, http.StatusBadRequest, "invalid fonte")
			return
		}
		fonte = &ft
	}

	obs, err := h.svc.ListObservacoes(r.Context(), skuID, fonte)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, obs)
}

func (h *Handlers) GetDelta(w http.ResponseWriter, r *http.Request) {
	skuID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	fonteStr := r.URL.Query().Get("fonte")
	if fonteStr == "" {
		writeError(w, http.StatusBadRequest, "fonte query param required")
		return
	}
	fonte := domain.Fonte(fonteStr)
	if !fonte.Valid() {
		writeError(w, http.StatusBadRequest, "invalid fonte")
		return
	}

	result, err := h.svc.GetDelta(r.Context(), skuID, fonte)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if result == nil {
		writeError(w, http.StatusNotFound, "no observations found")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handlers) GetMonitor(w http.ResponseWriter, r *http.Request) {
	skuID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	monitor, err := h.svc.GetActiveMonitor(r.Context(), skuID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if monitor == nil {
		writeError(w, http.StatusNotFound, "no active monitor")
		return
	}
	writeJSON(w, http.StatusOK, monitor)
}

func (h *Handlers) SetMonitor(w http.ResponseWriter, r *http.Request) {
	skuID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req domain.SetMonitorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.SKUID = skuID

	monitor, err := h.svc.SetActiveMonitor(r.Context(), &req)
	if err != nil {
		if errors.Is(err, domain.ErrSKUNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		var ve *domain.ValidationError
		if errors.As(err, &ve) || errors.Is(err, domain.ErrLimiarModoRequired) || errors.Is(err, domain.ErrInvalidLimiarModo) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, monitor)
}

func (h *Handlers) UpdateMonitorLimiar(w http.ResponseWriter, r *http.Request) {
	skuID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req domain.UpdateMonitorLimiarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	monitor, err := h.svc.UpdateMonitorLimiar(r.Context(), skuID, &req)
	if err != nil {
		if errors.Is(err, domain.ErrMonitorNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		var ve *domain.ValidationError
		if errors.As(err, &ve) || errors.Is(err, domain.ErrLimiarModoRequired) || errors.Is(err, domain.ErrInvalidLimiarModo) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, monitor)
}

func (h *Handlers) ListMonitorHistory(w http.ResponseWriter, r *http.Request) {
	skuID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	monitors, err := h.svc.ListMonitorHistory(r.Context(), skuID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, monitors)
}

func (h *Handlers) ListActiveMonitors(w http.ResponseWriter, r *http.Request) {
	monitors, err := h.svc.ListAllActiveMonitors(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, monitors)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
