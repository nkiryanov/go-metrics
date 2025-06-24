package app

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/capturer"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter/httpreporter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	halfSecond = 500 * time.Millisecond
)

func TestAgent_RunStoppedOnSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), halfSecond)
	defer cancel()

	// Prefer not to use mock here, cause it made test closer to production use
	agent := &Agent{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,

		Reporter: httpreporter.New("http://localhost:40010", &http.Client{}, nil, ""),
		Capturer: capturer.NewMemCapturer(),
	}

	err := agent.Run(ctx)

	require.Error(t, err)
	assert.Equal(t, "agent: Agent stopped", err.Error())
}
