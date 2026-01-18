package pprofserver

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
)

type PprofServer struct {
	listenAddr string
	lgr        logger.Logger
}

func New(listenAddr string, lgr logger.Logger) *PprofServer {
	return &PprofServer{
		listenAddr: listenAddr,
		lgr:        lgr,
	}
}

// Run starts http server with registered pprof handler and closes gracefully on context cancellation
func (s *PprofServer) Run(ctx context.Context) error {
	if s.listenAddr == "" {
		s.lgr.Debug("No pprof address, skip running pprof server")
		return nil
	}

	s.lgr.Info("Starting pprof server", "listenAddr", s.listenAddr)

	// Register pprof tools manually, as it does at net/http
	mux := http.NewServeMux()
	mux.HandleFunc("GET /debug/pprof/", pprof.Index)
	mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)

	httpServer := &http.Server{
		Addr:    s.listenAddr,
		Handler: mux,
	}

	idleConnsClosed := make(chan struct{})
	srvCtx, srvCtxCancel := context.WithCancel(ctx)
	defer srvCtxCancel()

	go func() {
		<-srvCtx.Done()

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := httpServer.Shutdown(timeoutCtx); err == context.DeadlineExceeded {
			s.lgr.Error("force pprof http server shutdown...")
		}
		s.lgr.Info("Pprof HTTP server stopped")
		close(idleConnsClosed)
	}()

	// Listen and serve until context is cancelled; then close gracefully connections
	err := httpServer.ListenAndServe()
	srvCtxCancel()
	<-idleConnsClosed

	return err
}
