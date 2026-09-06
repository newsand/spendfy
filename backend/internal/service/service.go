package service

import (
	"context"

	"github.com/newsand/spendfy/backend/internal/domain"
	"github.com/newsand/spendfy/backend/internal/repository"
)

type Service struct {
	repo *repository.PostgresRepository
}

func New(repo *repository.PostgresRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateSKU(ctx context.Context, req *domain.CreateSKURequest) (*domain.SKU, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return s.repo.CreateSKU(ctx, req)
}

func (s *Service) GetSKU(ctx context.Context, id int64) (*domain.SKU, error) {
	return s.repo.GetSKU(ctx, id)
}

func (s *Service) ListSKUs(ctx context.Context) ([]domain.SKU, error) {
	return s.repo.ListSKUs(ctx)
}

func (s *Service) UpdateSKU(ctx context.Context, id int64, req *domain.UpdateSKURequest) (*domain.SKU, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return s.repo.UpdateSKU(ctx, id, req)
}

func (s *Service) DeleteSKU(ctx context.Context, id int64) error {
	return s.repo.DeleteSKU(ctx, id)
}

func (s *Service) CreateObservacao(ctx context.Context, req *domain.CreateObservacaoRequest) (*domain.PrecoObservado, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	sku, err := s.repo.GetSKU(ctx, req.SKUID)
	if err != nil {
		return nil, err
	}

	if err := req.ValidateAgainstSKU(sku); err != nil {
		return nil, err
	}

	return s.repo.CreateObservacao(ctx, req)
}

func (s *Service) GetObservacao(ctx context.Context, id int64) (*domain.PrecoObservado, error) {
	return s.repo.GetObservacao(ctx, id)
}

func (s *Service) ListObservacoes(ctx context.Context, skuID int64, fonte *domain.Fonte) ([]domain.PrecoObservado, error) {
	return s.repo.ListObservacoesBySKU(ctx, skuID, fonte)
}

func (s *Service) GetDelta(ctx context.Context, skuID int64, fonte domain.Fonte) (*domain.DeltaResult, error) {
	var url *string
	if fonte == domain.FonteRastreio {
		monitor, err := s.repo.GetActiveMonitor(ctx, skuID)
		if err != nil {
			return nil, err
		}
		if monitor != nil {
			url = &monitor.URL
		}
	}

	current, err := s.repo.GetLastObservacaoByFonte(ctx, skuID, fonte, url)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}

	previous, err := s.repo.GetPreviousObservacao(ctx, skuID, fonte, current.ID, url)
	if err != nil {
		return nil, err
	}

	result := domain.CalculateDelta(current, previous)
	return &result, nil
}

func (s *Service) SetActiveMonitor(ctx context.Context, req *domain.SetMonitorRequest) (*domain.MonitorURL, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	_, err := s.repo.GetSKU(ctx, req.SKUID)
	if err != nil {
		return nil, err
	}

	return s.repo.SetActiveMonitor(ctx, req)
}

func (s *Service) GetActiveMonitor(ctx context.Context, skuID int64) (*domain.MonitorURL, error) {
	return s.repo.GetActiveMonitor(ctx, skuID)
}

func (s *Service) UpdateMonitorLimiar(ctx context.Context, skuID int64, req *domain.UpdateMonitorLimiarRequest) (*domain.MonitorURL, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return s.repo.UpdateMonitorLimiar(ctx, skuID, req)
}

func (s *Service) ListMonitorHistory(ctx context.Context, skuID int64) ([]domain.MonitorURL, error) {
	return s.repo.ListMonitorsBySKU(ctx, skuID, false)
}

func (s *Service) ListAllActiveMonitors(ctx context.Context) ([]domain.MonitorURL, error) {
	return s.repo.ListAllActiveMonitors(ctx)
}

func (s *Service) GetLastObservacaoForRastreio(ctx context.Context, skuID int64, url string) (*domain.PrecoObservado, error) {
	return s.repo.GetLastObservacaoByFonte(ctx, skuID, domain.FonteRastreio, &url)
}

func (s *Service) LogColeta(ctx context.Context, monitorID int64, success bool, obsID *int64, errMsg *string) error {
	return s.repo.LogColeta(ctx, monitorID, success, obsID, errMsg)
}

func (s *Service) ListTrackedSKUs(ctx context.Context) ([]domain.TrackedSKU, error) {
	return s.repo.ListTrackedSKUs(ctx)
}
