package main

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nkiryanov/go-metrics/cmd/server/app"
	"github.com/nkiryanov/go-metrics/internal/handlers"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

const (
	ListenAddr     string = ":8080"
	metricsListTpl string = "internal/handlers/templates/metrics_list.html"
)

var srv *app.ServerApp

func init() {
	listTpl, err := template.ParseFiles(metricsListTpl)

	if err != nil {
		panic(err)
	}

	api := handlers.NewMetricsAPI(storage.NewMemStorage(), listTpl)

	srv = &app.ServerApp{
		ListenAddr: ListenAddr,
		API:        api,
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		slog.Warn("Interrupt signal")
		cancel()
	}()

	if err := srv.Run(ctx); err != http.ErrServerClosed {
		slog.Error("HTTP server Shutdown", "error", err.Error())
		os.Exit(1)
	}
}
