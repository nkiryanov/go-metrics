package server

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	HalfSecond time.Duration = 500 * time.Millisecond
)

func TestMetricsServer_RunExitWithSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), HalfSecond)
	defer cancel()
	srv := MetricsServer{ListenAddr: "localhost:61999"}

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}
