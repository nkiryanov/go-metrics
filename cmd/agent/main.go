package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nkiryanov/go-metrics/cmd/agent/app"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

const (
	PollInterval = 2 * time.Second
	PubInterval  = 10 * time.Second

	PubAddr = "http://localhost:8080"
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
	agent := app.NewAgent(storage, PubAddr, PollInterval, PubInterval)

	if err := agent.Run(ctx); err != app.ErrAgentStopped {
		slog.Error("Something terrible happened", "error", err)
	}
}
