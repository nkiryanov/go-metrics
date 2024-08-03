package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nkiryanov/go-metrics/cmd/agent/app"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		slog.Warn("Interrupt signal")
		cancel()
	}()

	storage := storage.NewMemStorage()
	agent := app.NewAgent(storage)

	if err := agent.Poll(ctx); err != app.ErrAgentStopped {
		slog.Error("Something terrible happened", "error", err)
	}
}
