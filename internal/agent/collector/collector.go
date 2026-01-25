package collector

import (
	"time"

	"context"

	"github.com/nkiryanov/go-metrics/internal/models"
)

//go:generate go tool moq -out mocks/collector.go -pkg mocks -skip-ensure -fmt goimports . Collector

// Collector collects system metrics at configured intervals, stores them internally
type Collector interface {
	// Collect metrics. Should store them internally
	Collect(ctx context.Context) error

	// List collected metrics. Must be fast enough
	List() []models.Metric
}

// Run Collector periodically by 'interval'
func Run(ctx context.Context, c Collector, interval time.Duration) error {
	var err error

	err = c.Collect(ctx) // collect metrics at start
	if err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err = c.Collect(ctx)
			if err != nil {
				return err
			}
		}
	}
}
