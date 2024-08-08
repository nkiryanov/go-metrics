package runner

import (
	"context"
	"time"
)

func (p HTTPReporter) Run(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(p.repInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.batchReport()
				slog.Info("Metrics published")
				ticker.Reset(p.repInterval)
			}
		}
	}()

	<-ctx.Done()

	return ErrReporterStopped
}
