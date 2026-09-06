package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/newsand/spendfy/backend/internal/domain"
	"github.com/shopspring/decimal"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) CreateSKU(ctx context.Context, req *domain.CreateSKURequest) (*domain.SKU, error) {
	var sku domain.SKU
	err := r.pool.QueryRow(ctx, `
		INSERT INTO skus (nome, marca, tamanho_ou_variante, unidade_padrao, chave_identidade)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, nome, marca, tamanho_ou_variante, unidade_padrao, chave_identidade, created_at, updated_at
	`, req.Nome, req.Marca, req.TamanhoVariante, req.UnidadePadrao, req.ChaveIdentidade).Scan(
		&sku.ID, &sku.Nome, &sku.Marca, &sku.TamanhoVariante, &sku.UnidadePadrao, &sku.ChaveIdentidade, &sku.CreatedAt, &sku.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrChaveIdentidadeExists
		}
		return nil, fmt.Errorf("create sku: %w", err)
	}
	return &sku, nil
}

func (r *PostgresRepository) GetSKU(ctx context.Context, id int64) (*domain.SKU, error) {
	var sku domain.SKU
	err := r.pool.QueryRow(ctx, `
		SELECT id, nome, marca, tamanho_ou_variante, unidade_padrao, chave_identidade, created_at, updated_at
		FROM skus WHERE id = $1
	`, id).Scan(
		&sku.ID, &sku.Nome, &sku.Marca, &sku.TamanhoVariante, &sku.UnidadePadrao, &sku.ChaveIdentidade, &sku.CreatedAt, &sku.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSKUNotFound
		}
		return nil, fmt.Errorf("get sku: %w", err)
	}
	return &sku, nil
}

func (r *PostgresRepository) ListSKUs(ctx context.Context) ([]domain.SKU, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, nome, marca, tamanho_ou_variante, unidade_padrao, chave_identidade, created_at, updated_at
		FROM skus ORDER BY nome
	`)
	if err != nil {
		return nil, fmt.Errorf("list skus: %w", err)
	}
	defer rows.Close()

	var skus []domain.SKU
	for rows.Next() {
		var sku domain.SKU
		if err := rows.Scan(&sku.ID, &sku.Nome, &sku.Marca, &sku.TamanhoVariante, &sku.UnidadePadrao, &sku.ChaveIdentidade, &sku.CreatedAt, &sku.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan sku: %w", err)
		}
		skus = append(skus, sku)
	}
	return skus, rows.Err()
}

func (r *PostgresRepository) UpdateSKU(ctx context.Context, id int64, req *domain.UpdateSKURequest) (*domain.SKU, error) {
	var sku domain.SKU
	err := r.pool.QueryRow(ctx, `
		UPDATE skus SET
			nome = COALESCE($2, nome),
			marca = COALESCE($3, marca),
			tamanho_ou_variante = COALESCE($4, tamanho_ou_variante),
			unidade_padrao = COALESCE($5, unidade_padrao),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, nome, marca, tamanho_ou_variante, unidade_padrao, chave_identidade, created_at, updated_at
	`, id, req.Nome, req.Marca, req.TamanhoVariante, req.UnidadePadrao).Scan(
		&sku.ID, &sku.Nome, &sku.Marca, &sku.TamanhoVariante, &sku.UnidadePadrao, &sku.ChaveIdentidade, &sku.CreatedAt, &sku.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSKUNotFound
		}
		return nil, fmt.Errorf("update sku: %w", err)
	}
	return &sku, nil
}

func (r *PostgresRepository) DeleteSKU(ctx context.Context, id int64) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM skus WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete sku: %w", err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrSKUNotFound
	}
	return nil
}

func (r *PostgresRepository) CreateObservacao(ctx context.Context, req *domain.CreateObservacaoRequest) (*domain.PrecoObservado, error) {
	var obs domain.PrecoObservado
	err := r.pool.QueryRow(ctx, `
		INSERT INTO precos_observados (sku_id, preco, quantidade, unidade, loja, data, fonte, url, notas)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, sku_id, preco, moeda, quantidade, unidade, loja, data, fonte, url, notas, created_at
	`, req.SKUID, req.Preco, req.Quantidade, req.Unidade, req.Loja, time.Now().UTC(), req.Fonte, req.URL, req.Notas).Scan(
		&obs.ID, &obs.SKUID, &obs.Preco, &obs.Moeda, &obs.Quantidade, &obs.Unidade, &obs.Loja, &obs.Data, &obs.Fonte, &obs.URL, &obs.Notas, &obs.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create observacao: %w", err)
	}
	return &obs, nil
}

func (r *PostgresRepository) GetObservacao(ctx context.Context, id int64) (*domain.PrecoObservado, error) {
	var obs domain.PrecoObservado
	err := r.pool.QueryRow(ctx, `
		SELECT id, sku_id, preco, moeda, quantidade, unidade, loja, data, fonte, url, notas, created_at
		FROM precos_observados WHERE id = $1
	`, id).Scan(
		&obs.ID, &obs.SKUID, &obs.Preco, &obs.Moeda, &obs.Quantidade, &obs.Unidade, &obs.Loja, &obs.Data, &obs.Fonte, &obs.URL, &obs.Notas, &obs.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrObservacaoNotFound
		}
		return nil, fmt.Errorf("get observacao: %w", err)
	}
	return &obs, nil
}

func (r *PostgresRepository) ListObservacoesBySKU(ctx context.Context, skuID int64, fonte *domain.Fonte) ([]domain.PrecoObservado, error) {
	query := `
		SELECT id, sku_id, preco, moeda, quantidade, unidade, loja, data, fonte, url, notas, created_at
		FROM precos_observados
		WHERE sku_id = $1
	`
	args := []any{skuID}

	if fonte != nil {
		query += ` AND fonte = $2`
		args = append(args, *fonte)
	}
	query += ` ORDER BY data DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list observacoes: %w", err)
	}
	defer rows.Close()

	var observacoes []domain.PrecoObservado
	for rows.Next() {
		var obs domain.PrecoObservado
		if err := rows.Scan(&obs.ID, &obs.SKUID, &obs.Preco, &obs.Moeda, &obs.Quantidade, &obs.Unidade, &obs.Loja, &obs.Data, &obs.Fonte, &obs.URL, &obs.Notas, &obs.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan observacao: %w", err)
		}
		observacoes = append(observacoes, obs)
	}
	return observacoes, rows.Err()
}

func (r *PostgresRepository) GetPreviousObservacao(ctx context.Context, skuID int64, fonte domain.Fonte, currentID int64, url *string) (*domain.PrecoObservado, error) {
	var obs domain.PrecoObservado
	var query string
	var args []any

	if fonte == domain.FonteRastreio && url != nil {
		query = `
			SELECT id, sku_id, preco, moeda, quantidade, unidade, loja, data, fonte, url, notas, created_at
			FROM precos_observados
			WHERE sku_id = $1 AND fonte = $2 AND url = $3 AND id != $4
			ORDER BY data DESC LIMIT 1
		`
		args = []any{skuID, fonte, *url, currentID}
	} else {
		query = `
			SELECT id, sku_id, preco, moeda, quantidade, unidade, loja, data, fonte, url, notas, created_at
			FROM precos_observados
			WHERE sku_id = $1 AND fonte = $2 AND id != $3
			ORDER BY data DESC LIMIT 1
		`
		args = []any{skuID, fonte, currentID}
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&obs.ID, &obs.SKUID, &obs.Preco, &obs.Moeda, &obs.Quantidade, &obs.Unidade, &obs.Loja, &obs.Data, &obs.Fonte, &obs.URL, &obs.Notas, &obs.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get previous observacao: %w", err)
	}
	return &obs, nil
}

func (r *PostgresRepository) GetLastObservacaoByFonte(ctx context.Context, skuID int64, fonte domain.Fonte, url *string) (*domain.PrecoObservado, error) {
	var obs domain.PrecoObservado
	var query string
	var args []any

	if fonte == domain.FonteRastreio && url != nil {
		query = `
			SELECT id, sku_id, preco, moeda, quantidade, unidade, loja, data, fonte, url, notas, created_at
			FROM precos_observados
			WHERE sku_id = $1 AND fonte = $2 AND url = $3
			ORDER BY data DESC LIMIT 1
		`
		args = []any{skuID, fonte, *url}
	} else {
		query = `
			SELECT id, sku_id, preco, moeda, quantidade, unidade, loja, data, fonte, url, notas, created_at
			FROM precos_observados
			WHERE sku_id = $1 AND fonte = $2
			ORDER BY data DESC LIMIT 1
		`
		args = []any{skuID, fonte}
	}

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&obs.ID, &obs.SKUID, &obs.Preco, &obs.Moeda, &obs.Quantidade, &obs.Unidade, &obs.Loja, &obs.Data, &obs.Fonte, &obs.URL, &obs.Notas, &obs.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get last observacao: %w", err)
	}
	return &obs, nil
}

func (r *PostgresRepository) SetActiveMonitor(ctx context.Context, req *domain.SetMonitorRequest) (*domain.MonitorURL, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		UPDATE monitor_urls SET ativo = FALSE, archived_at = NOW()
		WHERE sku_id = $1 AND ativo = TRUE
	`, req.SKUID)
	if err != nil {
		return nil, fmt.Errorf("archive previous monitor: %w", err)
	}

	var monitor domain.MonitorURL
	err = tx.QueryRow(ctx, `
		INSERT INTO monitor_urls (sku_id, url, ativo, limiar_modo, limiar_valor, css_selector, regex_pattern)
		VALUES ($1, $2, TRUE, $3, $4, $5, $6)
		RETURNING id, sku_id, url, ativo, limiar_modo, limiar_valor, css_selector, regex_pattern, created_at, archived_at
	`, req.SKUID, req.URL, req.LimiarModo, req.LimiarValor, req.CSSSelector, req.RegexPattern).Scan(
		&monitor.ID, &monitor.SKUID, &monitor.URL, &monitor.Ativo, &monitor.LimiarModo, &monitor.LimiarValor,
		&monitor.CSSSelector, &monitor.RegexPattern, &monitor.CreatedAt, &monitor.ArchivedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create monitor: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &monitor, nil
}

func (r *PostgresRepository) GetActiveMonitor(ctx context.Context, skuID int64) (*domain.MonitorURL, error) {
	var monitor domain.MonitorURL
	err := r.pool.QueryRow(ctx, `
		SELECT id, sku_id, url, ativo, limiar_modo, limiar_valor, css_selector, regex_pattern, created_at, archived_at
		FROM monitor_urls
		WHERE sku_id = $1 AND ativo = TRUE
	`, skuID).Scan(
		&monitor.ID, &monitor.SKUID, &monitor.URL, &monitor.Ativo, &monitor.LimiarModo, &monitor.LimiarValor,
		&monitor.CSSSelector, &monitor.RegexPattern, &monitor.CreatedAt, &monitor.ArchivedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get active monitor: %w", err)
	}
	return &monitor, nil
}

func (r *PostgresRepository) UpdateMonitorLimiar(ctx context.Context, skuID int64, req *domain.UpdateMonitorLimiarRequest) (*domain.MonitorURL, error) {
	var monitor domain.MonitorURL
	err := r.pool.QueryRow(ctx, `
		UPDATE monitor_urls SET
			limiar_modo = $2,
			limiar_valor = $3
		WHERE sku_id = $1 AND ativo = TRUE
		RETURNING id, sku_id, url, ativo, limiar_modo, limiar_valor, css_selector, regex_pattern, created_at, archived_at
	`, skuID, req.LimiarModo, req.LimiarValor).Scan(
		&monitor.ID, &monitor.SKUID, &monitor.URL, &monitor.Ativo, &monitor.LimiarModo, &monitor.LimiarValor,
		&monitor.CSSSelector, &monitor.RegexPattern, &monitor.CreatedAt, &monitor.ArchivedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMonitorNotFound
		}
		return nil, fmt.Errorf("update monitor limiar: %w", err)
	}
	return &monitor, nil
}

func (r *PostgresRepository) ListMonitorsBySKU(ctx context.Context, skuID int64, activeOnly bool) ([]domain.MonitorURL, error) {
	query := `
		SELECT id, sku_id, url, ativo, limiar_modo, limiar_valor, css_selector, regex_pattern, created_at, archived_at
		FROM monitor_urls
		WHERE sku_id = $1
	`
	if activeOnly {
		query += ` AND ativo = TRUE`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, skuID)
	if err != nil {
		return nil, fmt.Errorf("list monitors: %w", err)
	}
	defer rows.Close()

	var monitors []domain.MonitorURL
	for rows.Next() {
		var m domain.MonitorURL
		if err := rows.Scan(&m.ID, &m.SKUID, &m.URL, &m.Ativo, &m.LimiarModo, &m.LimiarValor, &m.CSSSelector, &m.RegexPattern, &m.CreatedAt, &m.ArchivedAt); err != nil {
			return nil, fmt.Errorf("scan monitor: %w", err)
		}
		monitors = append(monitors, m)
	}
	return monitors, rows.Err()
}

func (r *PostgresRepository) ListAllActiveMonitors(ctx context.Context) ([]domain.MonitorURL, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, sku_id, url, ativo, limiar_modo, limiar_valor, css_selector, regex_pattern, created_at, archived_at
		FROM monitor_urls
		WHERE ativo = TRUE
		ORDER BY sku_id
	`)
	if err != nil {
		return nil, fmt.Errorf("list all monitors: %w", err)
	}
	defer rows.Close()

	var monitors []domain.MonitorURL
	for rows.Next() {
		var m domain.MonitorURL
		if err := rows.Scan(&m.ID, &m.SKUID, &m.URL, &m.Ativo, &m.LimiarModo, &m.LimiarValor, &m.CSSSelector, &m.RegexPattern, &m.CreatedAt, &m.ArchivedAt); err != nil {
			return nil, fmt.Errorf("scan monitor: %w", err)
		}
		monitors = append(monitors, m)
	}
	return monitors, rows.Err()
}

func (r *PostgresRepository) LogColeta(ctx context.Context, monitorID int64, success bool, obsID *int64, errMsg *string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO coleta_logs (monitor_id, success, preco_observado_id, error_message)
		VALUES ($1, $2, $3, $4)
	`, monitorID, success, obsID, errMsg)
	if err != nil {
		return fmt.Errorf("log coleta: %w", err)
	}
	return nil
}

type ColetaLog struct {
	ID               int64      `json:"id"`
	MonitorID        int64      `json:"monitor_id"`
	Success          bool       `json:"success"`
	PrecoObservadoID *int64     `json:"preco_observado_id,omitempty"`
	ErrorMessage     *string    `json:"error_message,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

func (r *PostgresRepository) ListColetaLogs(ctx context.Context, monitorID int64, limit int) ([]ColetaLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, monitor_id, success, preco_observado_id, error_message, created_at
		FROM coleta_logs
		WHERE monitor_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, monitorID, limit)
	if err != nil {
		return nil, fmt.Errorf("list coleta logs: %w", err)
	}
	defer rows.Close()

	var logs []ColetaLog
	for rows.Next() {
		var l ColetaLog
		if err := rows.Scan(&l.ID, &l.MonitorID, &l.Success, &l.PrecoObservadoID, &l.ErrorMessage, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan coleta log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "unique constraint"))
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

func (r *PostgresRepository) ListTrackedSKUs(ctx context.Context) ([]domain.TrackedSKU, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT s.id, s.nome, s.marca, s.tamanho_ou_variante, s.unidade_padrao, s.chave_identidade, s.created_at, s.updated_at
		FROM skus s
		LEFT JOIN precos_observados po ON s.id = po.sku_id
		LEFT JOIN monitor_urls mu ON s.id = mu.sku_id AND mu.ativo = TRUE
		WHERE po.id IS NOT NULL OR mu.id IS NOT NULL
		ORDER BY s.nome
	`)
	if err != nil {
		return nil, fmt.Errorf("list tracked skus: %w", err)
	}
	defer rows.Close()

	var result []domain.TrackedSKU
	for rows.Next() {
		var sku domain.SKU
		if err := rows.Scan(&sku.ID, &sku.Nome, &sku.Marca, &sku.TamanhoVariante, &sku.UnidadePadrao, &sku.ChaveIdentidade, &sku.CreatedAt, &sku.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan sku: %w", err)
		}
		result = append(result, domain.TrackedSKU{SKU: sku})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range result {
		skuID := result[i].SKU.ID

		var lastCompra *domain.PrecoObservado
		var lastCompraPrice decimal.Decimal
		var lastCompraData time.Time
		err := r.pool.QueryRow(ctx, `
			SELECT preco, data FROM precos_observados
			WHERE sku_id = $1 AND fonte = 'compra'
			ORDER BY data DESC LIMIT 1
		`, skuID).Scan(&lastCompraPrice, &lastCompraData)
		if err == nil {
			result[i].LastCompra = &lastCompraPrice
			result[i].LastCompraData = &lastCompraData
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get last compra for sku %d: %w", skuID, err)
		}
		_ = lastCompra

		var hasMonitor bool
		err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM monitor_urls WHERE sku_id = $1 AND ativo = TRUE)`, skuID).Scan(&hasMonitor)
		if err != nil {
			return nil, fmt.Errorf("check monitor for sku %d: %w", skuID, err)
		}
		result[i].HasMonitor = hasMonitor

		compraRows, err := r.pool.Query(ctx, `
			SELECT data, preco FROM precos_observados
			WHERE sku_id = $1 AND fonte = 'compra'
			ORDER BY data ASC
		`, skuID)
		if err != nil {
			return nil, fmt.Errorf("list compra series for sku %d: %w", skuID, err)
		}
		for compraRows.Next() {
			var pt domain.PricePoint
			if err := compraRows.Scan(&pt.Data, &pt.Preco); err != nil {
				compraRows.Close()
				return nil, fmt.Errorf("scan compra point: %w", err)
			}
			result[i].SeriesCompra = append(result[i].SeriesCompra, pt)
		}
		compraRows.Close()
		if err := compraRows.Err(); err != nil {
			return nil, err
		}

		rastreioRows, err := r.pool.Query(ctx, listRastreioSeriesForChartSQL, skuID)
		if err != nil {
			return nil, fmt.Errorf("list rastreio series for sku %d: %w", skuID, err)
		}
		for rastreioRows.Next() {
			var pt domain.PricePoint
			if err := rastreioRows.Scan(&pt.Data, &pt.Preco); err != nil {
				rastreioRows.Close()
				return nil, fmt.Errorf("scan rastreio point: %w", err)
			}
			result[i].SeriesRastreio = append(result[i].SeriesRastreio, pt)
		}
		rastreioRows.Close()
		if err := rastreioRows.Err(); err != nil {
			return nil, err
		}
	}

	return result, nil
}

var _ sql.Scanner = (*decimal.Decimal)(nil)
