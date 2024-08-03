package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/nkiryanov/go-metrics/internal/handlers"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

type ServerApp struct {
	ListenAddr string

	storage storage.Storage
	api     handlers.MetricsAPIHandler
}

func NewServerApp(listenAddr string) *ServerApp {
	storage := storage.NewMemStorage()
	return &ServerApp{
		ListenAddr: listenAddr,
		storage:    storage,
		api:        handlers.NewMetricsAPI(storage),
	}
}

func (s *ServerApp) router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/counter/{mName}/{mValue}", s.api.UpdateCounter)
	mux.HandleFunc("/update/gauge/{mName}/{mValue}", s.api.UpdateGauge)
	mux.HandleFunc("/update/{mType}/{mName}/{mValue}/", func(w http.ResponseWriter, r *http.Request) { http.Error(w, "Bad Request", http.StatusBadRequest) })
	return mux
}

// Run starts http server and closes gracefully on context cancellation
func (s *ServerApp) Run(ctx context.Context) error {
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
