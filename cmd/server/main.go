package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	_ "github.com/jackc/pgx/v5"

	"github.com/nkiryanov/go-metrics/cmd/server/opts"
	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/server/app"
	"github.com/nkiryanov/go-metrics/internal/server/handlers"
	"github.com/nkiryanov/go-metrics/internal/server/pprofserver"
	"github.com/nkiryanov/go-metrics/internal/server/storage"
	"github.com/nkiryanov/go-metrics/internal/server/storage/memstorage"
	"github.com/nkiryanov/go-metrics/internal/server/storage/pgstorage"
	"github.com/nkiryanov/go-metrics/internal/server/storage/pgstorage/db"
)

// Should be set on a build stage
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

// Defaults
const (
	listenAddr     = "localhost:8080"
	logLevel       = "info"
	dataFilePath   = "server_data.json"
	saveInterval   = 300 * time.Second
	restoreOnStart = false
	databaseDsn    = ""
	pprofAddr      = "" // empty by default = pprof disabled
)

func main() {
	opts := &opts.Options{
		ListenAddr:     listenAddr,
		LogLevel:       logLevel,
		DataFilePath:   dataFilePath,
		SaveInterval:   saveInterval,
		RestoreOnStart: restoreOnStart,
		DatabaseDsn:    databaseDsn,
		PprofAddr:      pprofAddr,
	}
	opts.Parse()

	lgr := logger.NewLogger(opts.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, opts, lgr); err != nil {
		lgr.Error("Server stopped with error", "error", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, opts *opts.Options, lgr logger.Logger) error {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	s, cleanup, err := initStorage(ctx, opts, lgr)
	if err != nil {
		return fmt.Errorf("storage initialization failed: %w", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			lgr.Error("Failed to cleanup storage", "error", err.Error())
		}
	}()

	g, gCtx := errgroup.WithContext(ctx)

	if opts.PprofAddr != "" {
		g.Go(func() error {
			pprofSrv := pprofserver.New(opts.PprofAddr, lgr)
			if err := pprofSrv.Run(gCtx); err != nil && err != http.ErrServerClosed {
				return fmt.Errorf("pprof server error: %w", err)
			}
			return nil
		})
	}

	g.Go(func() error {
		srv := app.New(opts.ListenAddr, handlers.NewMetricRouter(s, lgr, opts.SecretKey), lgr)
		if err := srv.Run(gCtx); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("HTTP server error: %w", err)
		}
		return nil
	})

	return g.Wait()
}

// initStorage initializes storage based on configuration.
// The returned cleanup func must be called to release resources — call it after all consumers have stopped.
func initStorage(ctx context.Context, opts *opts.Options, lgr logger.Logger) (storage.Storage, func() error, error) {
	switch {
	case opts.DatabaseDsn != "":
		// PostgreSQL storage: connect and run migrations before returning.
		dbpool, err := db.ConnectAndMigrate(ctx, opts.DatabaseDsn)
		if err != nil {
			return nil, nil, err
		}

		store := pgstorage.New(ctx, dbpool)
		cleanup := func() error {
			dbpool.Close()
			return store.Close()
		}
		return store, cleanup, nil

	default:
		// In-memory storage with optional file persistence.
		store, err := memstorage.New(opts.DataFilePath, opts.SaveInterval, opts.RestoreOnStart, lgr)
		if err != nil {
			return nil, nil, err
		}

		return store, store.Close, nil
	}
}
