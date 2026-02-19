package postgres

import (
	"database/sql"
	"errors"
	"time"

	"github.com/HyGames-fun/hytale-monitoring-backend/internal/domain"
	"github.com/jmoiron/sqlx"
)

type MetricsRepo struct {
	db *sqlx.DB
}

func NewMetricsRepo(db *sqlx.DB) *MetricsRepo {
	return &MetricsRepo{db: db}
}

func (r *MetricsRepo) SaveMetric(m *domain.ServerMetric) error {
	if m.RecordedAt.IsZero() {
		m.RecordedAt = time.Now()
	}
	query := `
	INSERT INTO "server_metrics" (server_id, is_online, players_online, players_max, tps, ram_used_mb, latency_ms, recorded_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING id;
	`
	return r.db.QueryRowx(query, m.ServerID, m.IsOnline, m.PlayersOnline, m.PlayersMax, m.TPS, m.RAMUsedMB, m.LatencyMS, m.RecordedAt).Scan(&m.ID)
}

func (r *MetricsRepo) GetLatestMetric(serverID int) (*domain.ServerMetric, error) {
	var m domain.ServerMetric
	query := `
	SELECT id, server_id, is_online, players_online, players_max, tps, ram_used_mb, latency_ms, recorded_at
	FROM "server_metrics"
	WHERE server_id = $1
	ORDER BY recorded_at DESC
	LIMIT 1
	`
	err := r.db.Get(&m, query, serverID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *MetricsRepo) GetLatestMetricsForServers(serverIDs []int) (map[int]*domain.ServerMetric, error) {
	result := make(map[int]*domain.ServerMetric)
	if len(serverIDs) == 0 {
		return result, nil
	}

	query, args, err := sqlx.In(`
		SELECT DISTINCT ON (server_id) id, server_id, is_online, players_online, players_max, tps, ram_used_mb, latency_ms, recorded_at
		FROM "server_metrics"
		WHERE server_id IN (?)
		ORDER BY server_id, recorded_at DESC
	`, serverIDs)
	if err != nil {
		return nil, err
	}

	query = r.db.Rebind(query)
	var metrics []domain.ServerMetric
	if err := r.db.Select(&metrics, query, args...); err != nil {
		return nil, err
	}

	for i := range metrics {
		result[metrics[i].ServerID] = &metrics[i]
	}
	return result, nil
}
