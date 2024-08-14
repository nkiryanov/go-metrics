package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/nkiryanov/go-metrics/cmd/server/opts"
)

type ServerApp struct {
	Opts    *opts.Options
	Handler http.Handler
}

// Run starts http server and closes gracefully on context cancellation
func (s *ServerApp) Run(ctx context.Context) error {
	slog.Info("Starting server", "ListenAddr", s.Opts.ListenAddr)

	httpServer := &http.Server{
		Addr:    s.Opts.ListenAddr,
		Handler: s.Handler,
	}

	idleConnsClosed := make(chan struct{})
	srvCtx, srvCtxCancel := context.WithCancel(ctx)
	defer srvCtxCancel()

	go func() {
		<-srvCtx.Done()

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(timeoutCtx); err == context.DeadlineExceeded {
			slog.Error("force http server shutdown...")
		}
		slog.Info("HTTP server stopped")
		close(idleConnsClosed)
	}()

	// Listen and serve until context is cancelled; then close gracefully connections
	err := httpServer.ListenAndServe()
	srvCtxCancel()
	<-idleConnsClosed

	return err
}
