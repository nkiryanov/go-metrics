package collector

import (
	"context"
	"github.com/nkiryanov/go-metrics/internal/models"
)

// Collector collects system metrics at configured intervals, stores them internally
type Collector interface {
	// Run starts the collector with its configured interval
	Run(ctx context.Context) error

	// List collected metrics. Should be fast enough
	List(ctx context.Context) ([]models.Metric, error)
}
