package app

import (
	"context"
	"errors"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/capturer"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter"
	"github.com/nkiryanov/go-metrics/internal/agent/runner"
)

var ErrAgentStopped = errors.New("agent: Agent stopped")

type Agent struct {
	PollIntv time.Duration
	ReptIntv time.Duration

	Rept reporter.Reporter
	Capt capturer.Capturer
}

func (a *Agent) Run(ctx context.Context) error {
	// create slice of all stored metrics and run report batch
	reportFn := func() {
		stats := a.Capt.Last()
		ms := make([]*reporter.Metric, 0, len(stats))

		for _, stat := range stats {
			ms = append(ms, &reporter.Metric{Name: stat.Name, Type: stat.Value.Type(), Value: stat.Value})
		}

		_ = a.Rept.ReportBatch(ms)
	}

	go runner.NewIntvRunner(0, a.PollIntv).Run(ctx, a.Capt.CaptureAndSave)
	go runner.NewIntvRunner(5*time.Second, a.ReptIntv).Run(ctx, reportFn)

	<-ctx.Done()
	return ErrAgentStopped
}
