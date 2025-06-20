package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/capturer"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter/httpreporter"
	"github.com/nkiryanov/go-metrics/internal/logger"

	"github.com/nkiryanov/go-metrics/cmd/agent/app"
	"github.com/nkiryanov/go-metrics/cmd/agent/opts"
)

const (
	ReportAddr = "http://localhost:8080"

	PollInterval   = 2 * time.Second
	ReportInterval = 10 * time.Second
	LogLevel       = "info"
	SecretKey      = ""
)

func main() {
	opts := &opts.Options{
		ReportAddr:     ReportAddr,
		PollInterval:   PollInterval,
		ReportInterval: ReportInterval,
		LogLevel:       LogLevel,
		SecretKey:      SecretKey,
	}
	opts.Parse()

	if err := logger.Initialize(opts.LogLevel); err != nil {
		log.Fatal("cant initialize logger, %w", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		logger.Slog.Warn("Interrupt signal")
		cancel()
	}()

	agent := &app.Agent{
		PollInterval:   opts.PollInterval,
		ReportInterval: opts.ReportInterval,

		Reporter: httpreporter.New(
			opts.ReportAddr,
			&http.Client{},
			[]time.Duration{time.Second, 3 * time.Second, 5 * time.Second},
			opts.SecretKey,
		),
		Capturer: capturer.NewMemCapturer(),
	}

	if err := agent.Run(ctx); err != app.ErrAgentStopped {
		logger.Slog.Error("Agent stopped unintentionally with error", "error", err)
		os.Exit(1)
	}
}
