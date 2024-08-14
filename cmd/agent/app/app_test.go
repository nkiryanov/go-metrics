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
	halfSecond = 500 * time.Millisecond
)

func TestAgent_RunStoppedOnSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), halfSecond)
	defer cancel()
	agent, _ := NewAgent(storage.NewMemStorage(), "http://localhost:101010", 2*time.Second, 10*time.Second)

	err := agent.Run(ctx)

	require.Error(t, err)
	assert.Equal(t, "agent: Agent stopped", err.Error())
}
