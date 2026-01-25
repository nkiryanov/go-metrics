package pprofserver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nkiryanov/go-metrics/internal/logger"
)

const (
	HalfSecond time.Duration = 500 * time.Millisecond
)

func TestPprofServer_ExitWithSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), HalfSecond)
	t.Cleanup(cancel)
	srv := New("localhost:49233", logger.NewNoOpLogger())

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}

func TestPprofServer_ExitOnServerError(t *testing.T) {
	ctx := context.Background()
	srv := New("199.23.23.999:8080", logger.NewNoOpLogger()) // Invalid address to trigger error

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listen")
}

func TestPprofServer_ExitOnEmptyAdder(t *testing.T) {
	ctx := t.Context()
	srv := New("", logger.NewNoOpLogger())

	err := srv.Run(ctx)

	require.NoError(t, err)
}
