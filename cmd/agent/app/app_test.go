package app

import (
	"context"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/capturer"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter"
	"github.com/nkiryanov/go-metrics/internal/storage"

	"github.com/go-resty/resty/v2"
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
		PollIntv: 2 * time.Second,
		ReptIntv: 10 * time.Second,

		Storage: storage.NewMemStorage(),
		Rept:    reporter.NewHTTPReporter("http://localhost:40010", resty.New()),
		Capt:    capturer.NewMemCapturer(),
	}

	err := agent.Run(ctx)

	require.Error(t, err)
	assert.Equal(t, "agent: Agent stopped", err.Error())
}
