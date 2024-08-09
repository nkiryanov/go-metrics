package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nkiryanov/go-metrics/internal/agent/capturer"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter"
	"github.com/nkiryanov/go-metrics/internal/storage"

	"github.com/nkiryanov/go-metrics/cmd/agent/app"
	"github.com/nkiryanov/go-metrics/cmd/agent/opts"
	"github.com/go-resty/resty/v2"
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

	agent := &app.Agent{
		PollIntv: opts.PollIntv,
		ReptIntv: opts.ReptIntv,

		Storage: storage.NewMemStorage(),
		Rept: reporter.NewHTTPReporter(opts.ReptAddr, resty.New()),
		Capt: capturer.NewMemCapturer(),
	}

	if err := agent.Run(ctx); err != app.ErrAgentStopped {
		slog.Error("Something terrible happened", "error", err)
		os.Exit(1)
	}
}
