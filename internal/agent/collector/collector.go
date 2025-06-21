package collector

import (
	"time"

	"context"
	"github.com/nkiryanov/go-metrics/internal/models"
)

// Collector collects system metrics at configured intervals, stores them internally
type Collector interface {
	// Collect metrics. Should store them internally
	Collect(ctx context.Context) error

	// List collected metrics. Should be fast enough
	List(ctx context.Context) ([]models.Metric, error)

}


// Run Collector periodically by 'interval'
func Run(ctx context.Context, c Collector, interval time.Duration) error {
	var err error

	err = c.Collect(ctx)  // collect metrics at start
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

