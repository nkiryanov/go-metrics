package main

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	"github.com/nkiryanov/go-metrics/internal/storage/memstorage"
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
	s, err := memstorage.New(opts.FilePath, opts.StoreInterval, opts.Restore)
	if err != nil {
		logger.Slog.Fatal("storage initialization failed", "error", err.Error())
	}
	defer s.Close()

	// Initialize database
	db, err := sql.Open("pgx", opts.Dsn)
	if err != nil {
		log.Fatal("couldn't initialize DB connection", err.Error())
	}
	defer db.Close()

	srv := &app.ServerApp{
		Opts:    opts,
		Handler: handlers.NewMetricRouter(s, db),
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
