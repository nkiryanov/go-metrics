package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/nkiryanov/go-metrics/internal/handlers"
)

type MetricsServer struct {
	ListenAddr string
	Runner     http.Handler
}

type Runner interface {
	Run() error
}

func (s *MetricsServer) router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{value}/", handlers.UpdateMetricHandler)
	return mux
}

// Run starts http server and closes gracefully on context cancellation
func (s *MetricsServer) Run(ctx context.Context) error {
	slog.Info("Starting server", "ListenAddr", s.ListenAddr)

	httpServer := &http.Server{
		Addr:    s.ListenAddr,
		Handler: s.router(),
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		<-ctx.Done()

		if httpServer != nil {
			timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := httpServer.Shutdown(timeoutCtx); err != nil {
				slog.Error("HTTP server Shutdown", "error", err.Error())
			}
			slog.Info("HTTP server stopped")
		}
		close(idleConnsClosed)
	}()

	// Listen and serve until context is cancelled
	err := httpServer.ListenAndServe()
	<-idleConnsClosed

	return err
}
