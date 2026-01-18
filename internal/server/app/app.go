package app

import (
	"context"
	"net/http"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
)

type ServerApp struct {
	ListenAddr string
	Handler    http.Handler
	lgr        logger.Logger
}

func New(listenAddr string, handler http.Handler, lgr logger.Logger) *ServerApp {
	return &ServerApp{
		ListenAddr: listenAddr,
		Handler:    handler,
		lgr:        lgr,
	}
}

// Run starts http server and closes gracefully on context cancellation
func (s *ServerApp) Run(ctx context.Context) error {
	s.lgr.Info("Starting server", "ListenAddr", s.ListenAddr)

	httpServer := &http.Server{
		Addr:    s.ListenAddr,
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
			s.lgr.Error("force http server shutdown...")
		}
		s.lgr.Info("HTTP server stopped")
		close(idleConnsClosed)
	}()

	// Listen and serve until context is cancelled; then close gracefully connections
	err := httpServer.ListenAndServe()
	srvCtxCancel()
	<-idleConnsClosed

	return err
}
