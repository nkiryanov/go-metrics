package storage

import (
	"github.com/nkiryanov/go-metrics/internal/models"
)

//go:generate moq -out mocks/storage.go -pkg mocks -skip-ensure -fmt goimports . Storage

type IterFunc func(models.Metric) error

type Storage interface {
	Count() int

	// Get metric from storage
	// If metric type is supported by the storage but 'mID' not found, then the 'ok' bool will be false.
	// If metric type is not supported by the storage, then 'err' will not be nil.
	GetMetric(mID string, mType string) (metric models.Metric, ok bool, err error)

	// Update metric in storage
	// May be implementation specific and not support all the types. In that case should return 'err'.
	UpdateMetric(in *models.Metric) (metric models.Metric, err error)

	// Iterate over stored values with 'iter' func.
	Iterate(iter IterFunc) error
}
