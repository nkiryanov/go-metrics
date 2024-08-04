package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/poller"
	"github.com/nkiryanov/go-metrics/internal/agent/publisher"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

var ErrAgentStopped = errors.New("agent: Agent stopped")

type Agent struct {
	storage   storage.Storage
	publisher publisher.Publisher
	poller    poller.Poller
}

func NewAgent(storage storage.Storage, pubAddr string, pollInterval, pubInterval time.Duration) (*Agent, error) {
	publisher, err := publisher.NewHTTPPublisher(pubAddr, pubInterval, storage)
	if err != nil {
		return nil, fmt.Errorf("agent: Failed to create publisher: %w", err)
	}

	return &Agent{
		storage:   storage,
		publisher: *publisher,
		poller:    poller.NewMemStatsPoller(storage, pollInterval),
	}, nil
}

func (a *Agent) Run(ctx context.Context) error {
	go a.poller.Run(ctx)    // nolint: errcheck
	go a.publisher.Run(ctx) // nolint: errcheck

	<-ctx.Done()
	return ErrAgentStopped
}
