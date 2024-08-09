package app

import (
	"context"
	"errors"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/capturer"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter"
	"github.com/nkiryanov/go-metrics/internal/agent/runner"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

var ErrAgentStopped = errors.New("agent: Agent stopped")

type Agent struct {
	PollIntv time.Duration
	ReptIntv time.Duration

	Rept    reporter.Reporter
	Capt    capturer.Capturer
	Storage storage.Storage
}

func (a *Agent) Run(ctx context.Context) error {
	// discard capture errors cause log them at corresponding packages
	captureFn := func() { _ = a.Capt.CaptureAndSave(a.Storage) }

	// create slice of all stored metrics and run report batch
	reportFn := func() {
		ms := make([]*reporter.Metric, 0, a.Storage.Len())
		a.Storage.Iterate(func(mType string, mName string, mValue storage.Storable) {
			ms = append(ms, &reporter.Metric{Name: mName, Type: mType, Value: mValue})
		})
		_ = a.Rept.ReportBatch(ms)
	}

	go runner.NewIntvRunner(0, a.PollIntv).Run(ctx, captureFn)
	go runner.NewIntvRunner(5*time.Second, a.ReptIntv).Run(ctx, reportFn)

	<-ctx.Done()
	return ErrAgentStopped
}
