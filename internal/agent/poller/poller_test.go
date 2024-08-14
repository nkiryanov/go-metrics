package poller

import (
	"context"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	HalfSecond = 500 * time.Millisecond
)

func TestAgent_PollStoppedOnSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), HalfSecond)
	defer cancel()
	poller := NewMemStatsPoller(storage.NewMemStorage(), time.Second)

	err := poller.Run(ctx)

	require.Error(t, err)
	assert.Equal(t, "poller: Poller stopped", err.Error())
}

func TestPoller_PollCaptureStats(t *testing.T) {
	poller := NewMemStatsPoller(storage.NewMemStorage(), time.Second)

	poller.captureStats()

	for _, gauge := range gauges {
		_, ok := poller.storage.GetGauge(gauge)
		assert.True(t, ok, "Expected gauge %s to be set", gauge)
	}
	for _, counter := range []storage.MetricName{PollCount} {
		_, ok := poller.storage.GetCounter(counter)
		assert.True(t, ok, "Expected counter %s to be set", counter)
	}
}
