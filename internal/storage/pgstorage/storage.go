package pgstorage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/storage/pgstorage/db"
	"github.com/nkiryanov/go-metrics/internal/storage/pgstorage/queries"
)

type PgStorage struct {
	db *pgxpool.Pool
}

// Create PgStorage
// Note: it embed migrations files, that would be run on initialization
func New(ctx context.Context, connString string) (*PgStorage, error) {
	pool, err := db.New(ctx, connString)
	return &PgStorage{db: pool}, err
}

func (s *PgStorage) Close() error {
	s.db.Close()
	return nil
}

func (s *PgStorage) Ping(ctx context.Context) error {
	return s.db.Ping(ctx)
}

func (s *PgStorage) CountMetric(ctx context.Context) (int, error) {
	q := queries.New(s.db)
	return q.CountMetric(ctx)
}

func (s *PgStorage) GetMetric(ctx context.Context, mType string, mName string) (models.Metric, error) {
	q := queries.New(s.db)
	return q.GetMetric(ctx, mType, mName)
}

func (s *PgStorage) UpdateMetric(ctx context.Context, m *models.Metric) (models.Metric, error) {
	q := queries.New(s.db)
	return q.UpdateMetric(ctx, m)
}

func (s *PgStorage) ListMetric(ctx context.Context) ([]models.Metric, error) {
	q := queries.New(s.db)
	return q.ListMetric(ctx)
}

// Get slice of metrics, update them in transaction and return slice of updated metrics
// Return err and rollback if any error occurs
func (s *PgStorage) UpdateMetricBulk(ctx context.Context, metrics []models.Metric) ([]models.Metric, error) {
	updated := make([]models.Metric, 0, len(metrics))
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		qtx := queries.WithTx(tx)
		var err error
		var u models.Metric
		for _, m := range metrics {
			if u, err = qtx.UpdateMetric(ctx, &m); err != nil {
				return err
			}
			updated = append(updated, u)
		}
		return nil
	})

	return updated, err
}
