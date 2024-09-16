package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nkiryanov/go-metrics/cmd/server/app"
	"github.com/nkiryanov/go-metrics/cmd/server/opts"
	"github.com/nkiryanov/go-metrics/internal/handlers"
	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

// Defaults
const (
	listenAddr    = "localhost:8080"
	logLevel      = "info"
	filePath      = "server_data.json"
	storeInterval = 300 * time.Second
	restore       = false
)

func main() {
	opts := &opts.Options{
		ListenAddr:    listenAddr,
		LogLevel:      logLevel,
		FilePath:      filePath,
		StoreInterval: storeInterval,
		Restore:       restore,
	}
	opts.Parse()

	// Init logger
	if err := logger.Initialize(opts.LogLevel); err != nil {
		log.Fatal("logger could not be initialized, %w", err.Error())
	}

	// Initialize storage
	s, err := storage.NewMemStorage(opts.FilePath, opts.StoreInterval, opts.Restore)
	if err != nil {
		logger.Slog.Fatal("storage initialization failed", "error", err.Error())
	}

	srv := &app.ServerApp{
		Opts:    opts,
		Handler: handlers.NewMetricRouter(s),
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		logger.Slog.Warn("Interrupt signal")
		cancel()
	}()

	if err := srv.Run(ctx); err != http.ErrServerClosed {
		logger.Slog.Error("HTTP server Shutdown", "error", err.Error())
		os.Exit(1)
	}
}
