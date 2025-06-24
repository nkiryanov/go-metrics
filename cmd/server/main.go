package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5"

	"github.com/nkiryanov/go-metrics/cmd/server/app"
	"github.com/nkiryanov/go-metrics/cmd/server/opts"
	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/server/handlers"
	"github.com/nkiryanov/go-metrics/internal/server/storage"
	"github.com/nkiryanov/go-metrics/internal/server/storage/memstorage"
	"github.com/nkiryanov/go-metrics/internal/server/storage/pgstorage"
	"github.com/nkiryanov/go-metrics/internal/server/storage/pgstorage/db"
)

// Defaults
const (
	listenAddr     = "localhost:8080"
	logLevel       = "info"
	dataFilePath   = "server_data.json"
	saveInterval   = 300 * time.Second
	restoreOnStart = false
	databaseDsn    = ""
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	opts := &opts.Options{
		ListenAddr:     listenAddr,
		LogLevel:       logLevel,
		DataFilePath:   dataFilePath,
		SaveInterval:   saveInterval,
		RestoreOnStart: restoreOnStart,
		DatabaseDsn:    databaseDsn,
	}
	opts.Parse()

	// Initialize logger
	if err := logger.Initialize(opts.LogLevel); err != nil {
		return fmt.Errorf("logger could not be initialized: %w", err)
	}

	ctx := context.Background()

	// Initialize storage
	s, cancelFn, err := initStorage(ctx, opts)
	if err != nil {
		return fmt.Errorf("storage initialization failed: %w", err)
	}
	defer func() {
		err := cancelFn()
		if err != nil {
			logger.Slog.Error("Failed to cleanup storage", "error", err.Error())
		}
	}()

	// Initialize context that cancelled on SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		logger.Slog.Warn("Interrupt signal")
		cancel()
	}()

	// Run server
	srv := app.New(opts.ListenAddr, handlers.NewMetricRouter(s, opts.SecretKey))
	if err := srv.Run(ctx); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

// Initializes storage based on configuration.
// Returns PostgreSQL storage if DSN is provided, otherwise returns memory storage with optional file persistence.
// The returned cancelFunc must be called to properly cleanup resources.
func initStorage(ctx context.Context, opts *opts.Options) (s storage.Storage, cancelFunc func() error, err error) {
	switch {
	case opts.DatabaseDsn != "":
		dbpool, err := db.ConnectAndMigrate(ctx, opts.DatabaseDsn)
		if err != nil {
			return nil, nil, err
		}

		pgs := pgstorage.New(ctx, dbpool)
		cancelFn := func() error {
			dbpool.Close()
			return pgs.Close()
		}
		return pgs, cancelFn, nil
	default:
		mems, err := memstorage.New(opts.DataFilePath, opts.SaveInterval, opts.RestoreOnStart)
		if err != nil {
			return nil, nil, err
		}

		return mems, mems.Close, nil
	}
}
