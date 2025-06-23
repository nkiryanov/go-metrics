package storage

import (
	"context"
	"errors"

	"github.com/nkiryanov/go-metrics/internal/models"
)

var (
	// ErrNoMetric occurs when no metric found, while expects.
	ErrNoMetric = errors.New("no metric found")
)

//go:generate moq -out mocks/storage.go -pkg mocks -skip-ensure -fmt goimports . Storage

type Storage interface {
	// Ping wether storage ready to work
	Ping(ctx context.Context) error

	// Count all the stored metrics
	CountMetric(ctx context.Context) (int, error)

	// Get metric from storage
	// If metric not found return errors.Is(ErrNoMetric), otherwise other storage access error
	GetMetric(ctx context.Context, mType string, mName string) (metric models.Metric, err error)

	// Update metric in storage
	// May be implementation specific and not support all the types. In that case should return 'err'.
	UpdateMetric(ctx context.Context, in models.Metric) (metric models.Metric, err error)

	// Update metrics bulk in storage
	UpdateMetricBulk(ctx context.Context, metrics []models.Metric) (updated []models.Metric, err error)

	// List all the metrics
	// Just an MVP: should return slice of metrics, ordered by Name
	ListMetric(ctx context.Context) (metrics []models.Metric, err error)

	// Close storage
	Close() error
}
