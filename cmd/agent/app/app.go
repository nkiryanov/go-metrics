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
	PollInterval   time.Duration
	ReportInterval time.Duration

	Reporter reporter.Reporter
	Capturer capturer.Capturer
}

func (a *Agent) Run(ctx context.Context) error {
	// create slice of all stored metrics and run report batch
	reportFn := func() { _ = a.Reporter.ReportBatch(a.Capturer.Last()) }

	go runner.NewIntvRunner(0, a.PollInterval).Run(ctx, a.Capturer.CaptureAndSave)
	go runner.NewIntvRunner(5*time.Second, a.ReportInterval).Run(ctx, reportFn)

	<-ctx.Done()
	return ErrAgentStopped
}
