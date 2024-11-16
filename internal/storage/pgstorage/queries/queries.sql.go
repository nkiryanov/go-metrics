package queries

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/storage"
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

const countMetric = `
	SELECT count(*) AS count
	FROM "metric"
	`

func (q *Queries) CountMetric(ctx context.Context) (int, error) {
	rows, _ := q.db.Query(ctx, countMetric)
	return pgx.CollectExactlyOneRow(rows, pgx.RowTo[int])
}

const getMetric = `
	SELECT "type", "name", "delta", "value"
	FROM "metric"
	WHERE "type" = $1 AND "name" = $2
	`

func (q *Queries) GetMetric(ctx context.Context, mType string, mName string) (models.Metric, error) {
	rows, _ := q.db.Query(ctx, getMetric, mType, mName)
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

func (q *Queries) UpdateMetric(ctx context.Context, m *models.Metric) (models.Metric, error) {
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

	rows, _ := q.db.Query(ctx, insertOrUpdateMetric, m.Type, m.Name, delta, value)
	return pgx.CollectExactlyOneRow(rows, rowToMetric)
}

const listMetric = `
	SELECT "type", "name", "delta", "value"
	FROM "metric"
	ORDER BY "name", "type"
	`

func (q *Queries) ListMetric(ctx context.Context) ([]models.Metric, error) {
	rows, _ := q.db.Query(ctx, listMetric)
	return pgx.CollectRows(rows, rowToMetric)
}
