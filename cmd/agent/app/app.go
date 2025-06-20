package app

import (
	"context"
	"errors"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/capturer"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter"
	"github.com/nkiryanov/go-metrics/internal/agent/runner"
	"github.com/nkiryanov/go-metrics/internal/logger"
)

var ErrAgentStopped = errors.New("agent: Agent stopped")

type Agent struct {
	PollInterval   time.Duration
	ReportInterval time.Duration

	Reporter reporter.Reporter
	Capturer capturer.Capturer
}

func (a *Agent) Run(ctx context.Context) error {
	// create slice of all stored metrics and run report batch
	reportFn := func() {
		captured := a.Capturer.ListLast()
		err := a.Reporter.ReportBatch(captured)

		switch err {
		case nil:
			logger.Slog.Info("metrics reported or", "count", len(captured))
		default:
			logger.Slog.Warn("report error", "error", err.Error())
		}
	}

	go runner.NewIntvRunner(0, a.PollInterval).Run(ctx, a.Capturer.CaptureAndSave)
	go runner.NewIntvRunner(5*time.Second, a.ReportInterval).Run(ctx, reportFn)

	<-ctx.Done()
	return ErrAgentStopped
}
