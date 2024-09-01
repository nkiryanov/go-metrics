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
	reportFn := func() { _ = a.Rept.ReportBatch(a.Capt.Last()) }

	go runner.NewIntvRunner(0, a.PollIntv).Run(ctx, a.Capt.CaptureAndSave)
	go runner.NewIntvRunner(5*time.Second, a.ReptIntv).Run(ctx, reportFn)

	<-ctx.Done()
	return ErrAgentStopped
}
