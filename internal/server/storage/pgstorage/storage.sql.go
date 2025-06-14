package pgstorage

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/server/storage"
)

func rowToMetric(row pgx.CollectableRow) (models.Metric, error) {
	var m models.Metric
	var delta pgtype.Int8
	var value pgtype.Float8
	var err error

	if err = row.Scan(&m.Type, &m.Name, &delta, &value); err != nil {
		return m, err
	}

	m.Delta = delta.Int64
	m.Value = value.Float64
	return m, nil
}

// Execute empty sql statement and return result
func (s *PgStorage) Ping(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `-- ping`)
	return err
}

const countMetric = `
SELECT count(*) AS count
FROM "metric"
`

func (s *PgStorage) CountMetric(ctx context.Context) (int, error) {
	rows, _ := s.db.Query(ctx, countMetric)
	return pgx.CollectExactlyOneRow(rows, pgx.RowTo[int])
}

const getMetric = `
SELECT "type", "name", "delta", "value"
FROM "metric"
WHERE "type" = $1 AND "name" = $2
`

func (s *PgStorage) GetMetric(ctx context.Context, mType string, mName string) (models.Metric, error) {
	rows, _ := s.db.Query(ctx, getMetric, mType, mName)
	metric, err := pgx.CollectExactlyOneRow(rows, rowToMetric)

	if errors.Is(err, pgx.ErrNoRows) {
		err = storage.ErrNoMetric
	}

	return metric, err
}

const insertOrUpdateMetric = `
INSERT INTO "metric" ("type", "name", "delta", "value")
VALUES ($1, $2, $3, $4)
ON CONFLICT ("name", "type")
DO UPDATE SET
	"delta" = "metric"."delta" + EXCLUDED."delta",
	"value" = EXCLUDED."value"
RETURNING "type", "name", "delta", "value"
`

func (s *PgStorage) UpdateMetric(ctx context.Context, m *models.Metric) (models.Metric, error) {
	var delta pgtype.Int8
	var value pgtype.Float8

	switch m.Type {
	case models.CounterTypeName:
		delta.Int64 = m.Delta
		delta.Valid = true
	case models.GaugeTypeName:
		value.Float64 = m.Value
		value.Valid = true
	}

	rows, _ := s.db.Query(ctx, insertOrUpdateMetric, m.Type, m.Name, delta, value)
	return pgx.CollectExactlyOneRow(rows, rowToMetric)
}

const listMetric = `
SELECT "type", "name", "delta", "value"
FROM "metric"
ORDER BY "name", "type"
`

func (s *PgStorage) ListMetric(ctx context.Context) ([]models.Metric, error) {
	rows, _ := s.db.Query(ctx, listMetric)
	return pgx.CollectRows(rows, rowToMetric)
}

// Get slice of metrics, update them in transaction and return slice of updated metrics
// Return err and rollback if any error occurs
func (s *PgStorage) UpdateMetricBulk(ctx context.Context, metrics []models.Metric) ([]models.Metric, error) {
	updated := make([]models.Metric, 0, len(metrics))
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		storageTx := WithTx(tx)
		var err error
		var u models.Metric
		for _, m := range metrics {
			if u, err = storageTx.UpdateMetric(ctx, &m); err != nil {
				return err
			}
			updated = append(updated, u)
		}
		return nil
	})

	return updated, err
}
