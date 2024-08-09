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
	pollIntv time.Duration
	reptIntv time.Duration

	rept    reporter.Reporter
	capt    capturer.Capturer
	storage storage.Storage
}

func NewAgent(storage storage.Storage, reptAddr string, pollIntv, reptIntv time.Duration) *Agent {
	return &Agent{
		rept:     reporter.NewHTTPReporter(reptAddr),
		capt:     capturer.NewMemCapturer(),
		storage:  storage,
		pollIntv: pollIntv,
		reptIntv: reptIntv,
	}
}

func (a *Agent) capture() {
	stats := a.capt.Capture()

	for _, stat := range stats {
		if _, err := a.storage.UpdateMetric(stat.Name, stat.Value); err != nil {
			slog.Error("agent: cant't update storage metric", "error", err.Error())
		}
	}

	slog.Info("agent: mem stats saved")
}

func (a *Agent) report() {
	ms := make([]*reporter.Metric, 0, a.storage.Len())

	a.storage.Iterate(func(mType string, mName string, mValue storage.Storable) {
		ms = append(ms, &reporter.Metric{Name: mName, Type: mType, Value: mValue})
	})

	if errs := a.rept.ReportBatch(ms); len(errs) > 0 {
		slog.Warn("agent: can't report metrics", "count", len(errs), "error", errs[0].Error())
	} else {
		slog.Info("agent: metrics reported", "count", len(ms))
	}
}

func (a *Agent) Run(ctx context.Context) error {
	go runner.NewIntvRunner(0, a.pollIntv).Run(ctx, a.capture)
	go runner.NewIntvRunner(5*time.Second, a.reptIntv).Run(ctx, a.report)

	<-ctx.Done()
	return ErrAgentStopped
}
