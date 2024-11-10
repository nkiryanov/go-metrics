package storage

import (
	"errors"

	"github.com/nkiryanov/go-metrics/internal/models"
)

var (
	// ErrNoMetric occurs when no metric found, while expects.
	ErrNoMetric = errors.New("no metric found")
)

//go:generate moq -out mocks/storage.go -pkg mocks -skip-ensure -fmt goimports . Storage

type Storage interface {
	CountMetric() int

	// Get metric from storage
	// If metric not found return errors.Is(ErrNoMetric), otherwise other storage access error
	GetMetric(mType string, mName string) (metric models.Metric, err error)

	// Update metric in storage
	// May be implementation specific and not support all the types. In that case should return 'err'.
	UpdateMetric(in *models.Metric) (metric models.Metric, err error)

	// List all the metrics
	// Just an MVP: should return slice of metrics, ordered by Name
	ListMetric() (metrics []models.Metric, err error)
}
