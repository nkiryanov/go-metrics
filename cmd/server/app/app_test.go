package app

import (
	"context"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	HalfSecond time.Duration = 500 * time.Millisecond
)

type mockAPI struct{}

func (m mockAPI) RegisterRoutes(chi.Router) {
	// Registered
}

func TestServerApp_RunExitWithSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), HalfSecond)
	defer cancel()
	srv := ServerApp{ListenAddr: ":8080", API: mockAPI{}}

	err := srv.Run(ctx)

	require.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}
