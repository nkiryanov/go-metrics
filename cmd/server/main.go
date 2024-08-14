package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nkiryanov/go-metrics/cmd/server/app"
	"github.com/nkiryanov/go-metrics/cmd/server/opts"
	"github.com/nkiryanov/go-metrics/internal/handlers"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

const (
	listenAddr = "localhost:8080"
)

func main() {
	opts := &opts.Options{
		ListenAddr: listenAddr,
	}

	opts.Parse()

	s := storage.NewMemStorage()

	srv := &app.ServerApp{
		Opts:    opts,
		Handler: handlers.NewMetricRouter(s, storage.MemParser{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		slog.Warn("Interrupt signal")
		cancel()
	}()

	if err := srv.Run(ctx); err != http.ErrServerClosed {
		slog.Error("HTTP server Shutdown", "error", err.Error())
		os.Exit(1)
	}
}
