package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nkiryanov/go-metrics/cmd/server/app"
	"github.com/nkiryanov/go-metrics/cmd/server/opts"
	"github.com/nkiryanov/go-metrics/internal/handlers"
	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/storage"
	"github.com/nkiryanov/go-metrics/internal/storage/memstorage"
	"github.com/nkiryanov/go-metrics/internal/storage/pgstorage"
)

// Defaults
const (
	listenAddr    = "localhost:8080"
	logLevel      = "info"
	filePath      = "server_data.json"
	storeInterval = 300 * time.Second
	restore       = false
	dsn           = "postgres://go-metrics@localhost:5432/go-metrics"
)

func main() {
	opts := &opts.Options{
		ListenAddr:    listenAddr,
		LogLevel:      logLevel,
		FilePath:      filePath,
		StoreInterval: storeInterval,
		Restore:       restore,
		Dsn:           dsn,
	}
	opts.Parse()

	// Initialize logger
	if err := logger.Initialize(opts.LogLevel); err != nil {
		log.Fatal("logger could not be initialized, %w", err.Error())
	}

	// Initialize storage
	s, err := initStorage(opts)
	if err != nil {
		logger.Slog.Fatal("storage initialization failed", "error", err.Error())
	}
	defer s.Close()

	srv := &app.ServerApp{
		Opts:    opts,
		Handler: handlers.NewMetricRouter(s),
	}

	// Initialize context that cancelled on SIGTERM
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

func initStorage(opts *opts.Options) (storage.Storage, error) {
	if opts.Dsn != "" {
		return pgstorage.New(context.TODO(), opts.Dsn)
	}

	return memstorage.New(opts.FilePath, opts.StoreInterval, opts.Restore)
}
