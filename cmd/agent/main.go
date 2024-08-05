package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nkiryanov/go-metrics/cmd/agent/app"
	"github.com/nkiryanov/go-metrics/cmd/agent/opts"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

func main() {
	opts := opts.ParseOptions()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		slog.Warn("Interrupt signal")
		cancel()
	}()

	storage := storage.NewMemStorage()
	agent, err := app.NewAgent(storage, opts.ReportAddr, opts.PollInterval, opts.ReportInterval)
	if err != nil {
		slog.Error("Failed to create agent", "error", err)
		os.Exit(1)
	}

	if err := agent.Run(ctx); err != app.ErrAgentStopped {
		slog.Error("Something terrible happened", "error", err)
		os.Exit(1)
	}
}
