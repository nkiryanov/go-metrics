package metrics_server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/nkiryanov/go-metrics/internal/handlers"
)

type Server struct {
	ListenAddr string
	Runner     http.Handler
}

type Runner interface {
	Run() error
}

func (s *Server) router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/{metricType}/{metricName}/{value}", handlers.UpdateMetricHandler)
	return mux
}

func (s *Server) Run(ctx context.Context) error {
	slog.Info("[INFO] Starting server", "ListenAddr", s.ListenAddr)

	httpServer := &http.Server{
		Addr:    s.ListenAddr,
		Handler: s.router(),
	}

	go func() {
		<-ctx.Done()
		if httpServer != nil {
			if err := httpServer.Close(); err != nil {
				slog.Error("[ERROR] Error closing server", "error", err.Error())
			}
		}
	}()

	return httpServer.ListenAndServe()
}
