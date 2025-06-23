package app

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

func TestServerApp_ExitWithSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), HalfSecond)
	t.Cleanup(cancel)
	srv := ServerApp{ListenAddr: "localhost:39232", Handler: nil}

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}

func TestServerApp_ExitOnServerError(t *testing.T) {
	ctx := context.Background()
	srv := ServerApp{ListenAddr: "19.23.23.999:8080", Handler: nil}

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listen")
}
