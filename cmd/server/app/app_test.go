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
	opts := &opts.Options{ListenAddr: opts.NetAddress{Host: "localhost", Port: 39232}}
	srv := ServerApp{Opts: opts, Handler: nil}

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}

func TestServerApp_ExitOnServerError(t *testing.T) {
	ctx := context.Background()
	opts := &opts.Options{ListenAddr: opts.NetAddress{Host: "19.23.23.999", Port: 8080}} // invalid host
	srv := ServerApp{Opts: opts, Handler: nil}

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "listen")
}
