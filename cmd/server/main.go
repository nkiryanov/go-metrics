package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5"

	"github.com/nkiryanov/go-metrics/cmd/server/app"
	"github.com/nkiryanov/go-metrics/cmd/server/opts"
	"github.com/nkiryanov/go-metrics/internal/db"
	"github.com/nkiryanov/go-metrics/internal/server/handlers"
	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/server/storage"
	"github.com/nkiryanov/go-metrics/internal/server/storage/memstorage"
	"github.com/nkiryanov/go-metrics/internal/server/storage/pgstorage"
)

// Defaults
const (
	listenAddr    = "localhost:8080"
	logLevel      = "info"
	filePath      = "server_data.json"
	storeInterval = 300 * time.Second
	restore       = false
	dsn           = ""
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

	ctx := context.Background()

	// Initialize storage
	s, cancelFn, err := initStorage(ctx, opts)
	if err != nil {
		logger.Slog.Fatal("storage initialization failed", "error", err.Error())
	}
	defer cancelFn() // nolint:errcheck

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

func initStorage(ctx context.Context, opts *opts.Options) (s storage.Storage, cancelFunc func() error, err error) {
	if opts.Dsn != "" {
		err = db.Migrate(opts.Dsn)
		if err != nil {
			return nil, nil, err
		}

		pool, err := db.Connect(ctx, opts.Dsn)
		if err != nil {
			return nil, nil, err
		}

		pgs := pgstorage.New(ctx, pool)
		cancelFn := func() error {
			pool.Close()
			return pgs.Close()
		}
		return pgstorage.New(ctx, pool), cancelFn, nil
	}

	ms, err := memstorage.New(opts.FilePath, opts.StoreInterval, opts.Restore)
	if err != nil {
		return nil, nil, err
	}

	return ms, ms.Close, nil
}
