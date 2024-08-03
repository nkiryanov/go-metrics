package app

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
	agent := NewAgent(storage.NewMemStorage())

	err := agent.Poll(ctx)

	require.Error(t, err)
	assert.Equal(t, "agent: Agent stopped", err.Error())
}

func TestAgent_PollCaptureStats(t *testing.T) {
	agent := NewAgent(storage.NewMemStorage())
	ctx, cancel := context.WithTimeout(context.Background(), HalfSecond)
	defer cancel()

	err := agent.Poll(ctx)

	require.Error(t, err)
	for _, gauge := range gauges {
		_, ok := agent.storage.GetGauge(gauge)
		assert.True(t, ok, "Expected gauge %s to be set", gauge)
	}
	for _, counter := range []storage.MetricName{PollCount} {
		_, ok := agent.storage.GetCounter(counter)
		assert.True(t, ok, "Expected counter %s to be set", counter)
	}
}
