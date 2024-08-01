package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nkiryanov/go-metrics/cmd/server/metrics_server"
)

const (
	ListenAddr string = ":8080"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		slog.Warn("[WARN] Interrupt signal")
		cancel()
	}()

	srv := metrics_server.Server{ListenAddr: ListenAddr}

	if err := srv.Run(ctx); err != nil {
		slog.Error("[ERROR] Error", "error", err.Error())
	}
}
