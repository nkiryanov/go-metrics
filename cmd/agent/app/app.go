package app

import (
	"context"
	"errors"
	"log/slog"
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

func (a *Agent) report() {
	ms := make([]*reporter.Metric, 0, a.Storage.Len())

	a.Storage.Iterate(func(mType string, mName string, mValue storage.Storable) {
		ms = append(ms, &reporter.Metric{Name: mName, Type: mType, Value: mValue})
	})

	if errs := a.Rept.ReportBatch(ms); len(errs) > 0 {
		slog.Warn("agent: can't report metrics", "count", len(errs), "error", errs[0].Error())
	} else {
		slog.Info("agent: metrics reported", "count", len(ms))
	}
}

func (a *Agent) Run(ctx context.Context) error {
	// discard capture errors cause them logged and that enough
	captureFn := func() { _ = a.Capt.CaptureWithSave(a.Storage) }

	go runner.NewIntvRunner(0, a.PollIntv).Run(ctx, captureFn)
	go runner.NewIntvRunner(5*time.Second, a.ReptIntv).Run(ctx, a.report)

	<-ctx.Done()
	return ErrAgentStopped
}
