package app

import (
	"context"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	HalfSecond time.Duration = 500 * time.Millisecond
)

func TestServerApp_ExitWithSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), HalfSecond)
	t.Cleanup(cancel)
	srv := New("localhost:39232", nil, logger.NewNoOpLogger())

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}

func TestServerApp_ExitOnServerError(t *testing.T) {
	ctx := context.Background()
	srv := New("19.23.23.999:8080", nil, logger.NewNoOpLogger()) // Invalid address to trigger error

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listen")
}
