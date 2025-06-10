package pgstorage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/db"
	"github.com/nkiryanov/go-metrics/internal/storage/pgstorage/queries"
)

type PgStorage struct {
	
	db *pgxpool.Pool
}

func New(ctx context.Context, dbURI string) (*PgStorage, error) {
	var err error

	pool, err := db.Connect(ctx, dbURI)
	if err != nil {
		return nil, err
	}

	err = db.Migrate(dbURI)
	if err != nil {
		return nil, err
	}

	return &PgStorage{db: pool}, nil
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

func (s *PgStorage) UpdateMetricBulk(ctx context.Context, metrics []models.Metric) ([]models.Metric, error) {
	q := queries.New(s.db)
	return q.UpdateMetricBulk(ctx, metrics)
}
