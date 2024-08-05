package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/poller"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

var ErrAgentStopped = errors.New("agent: Agent stopped")

type Agent struct {
	storage  storage.Storage
	reporter reporter.Reporter
	poller   poller.Poller
}

func NewAgent(storage storage.Storage, reptAddr string, pollInterval, reptInterval time.Duration) (*Agent, error) {
	reporter, err := reporter.NewHTTPReporter(reptAddr, reptInterval, storage)
	if err != nil {
		return nil, fmt.Errorf("agent: Failed to create publisher: %w", err)
	}

	return &Agent{
		storage:  storage,
		reporter: *reporter,
		poller:   poller.NewMemStatsPoller(storage, pollInterval),
	}, nil
}

func (a *Agent) Run(ctx context.Context) error {
	go a.poller.Run(ctx)   // nolint: errcheck
	go a.reporter.Run(ctx) // nolint: errcheck

	<-ctx.Done()
	return ErrAgentStopped
}
