package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nkiryanov/go-metrics/cmd/server/server"
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
		slog.Warn("Interrupt signal")
		cancel()
	}()

	srv := server.NewServerApp(ListenAddr)

	if err := srv.Run(ctx); err != http.ErrServerClosed {
		slog.Error("HTTP server Shutdown", "error", err.Error())
	}
}
