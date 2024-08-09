package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nkiryanov/go-metrics/cmd/server/opts"
)

const (
	HalfSecond time.Duration = 500 * time.Millisecond
)

func TestServerApp_ExitWithSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), HalfSecond)
	defer cancel()
	srv := ServerApp{Opts: opts.NewOptions(), Handler: nil}

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}

func TestServerApp_ExitOnServerError(t *testing.T) {
	ctx := context.Background()
	srv := ServerApp{Opts: &opts.Options{ListenAddr: "invalid_add"}, Handler: nil}

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listen")
}
