package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/capturer"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter"
	"github.com/nkiryanov/go-metrics/internal/logger"

	"github.com/go-resty/resty/v2"
	"github.com/nkiryanov/go-metrics/cmd/agent/app"
	"github.com/nkiryanov/go-metrics/cmd/agent/opts"
)

const (
	ReptAddr = "http://localhost:8080"

	PollIntv = 2 * time.Second
	ReptIntv = 10 * time.Second
)

func main() {
	opts := &opts.Options{
		ReptAddr: ReptAddr,
		PollIntv: PollIntv,
		ReptIntv: ReptIntv,
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
		PollIntv: opts.PollIntv,
		ReptIntv: opts.ReptIntv,

		Rept: reporter.NewHTTPReporter(opts.ReptAddr, resty.New()),
		Capt: capturer.NewMemCapturer(),
	}

	if err := agent.Run(ctx); err != app.ErrAgentStopped {
		logger.Slog.Error("Something terrible happened", "error", err)
		os.Exit(1)
	}
}
