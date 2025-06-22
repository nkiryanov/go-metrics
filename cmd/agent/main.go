package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/config"
	"github.com/nkiryanov/go-metrics/internal/logger"
)

const (
	LogLevel = "info"

	ReportAddr     = "http://localhost:8080"
	ReportInterval = 10 * time.Second
	SecretKey      = ""

	CollectInterval = 2 * time.Second
)

func main() {
	cfg := &config.Config{
		LogLevel:        LogLevel,
		ReportAddr:      ReportAddr,
		ReportInterval:  ReportInterval,
		SecretKey:       SecretKey,
		CollectInterval: CollectInterval,
	}
	cfg.MustLoad()

	lgr := logger.NewLogger(cfg.LogLevel)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		lgr.Warn("Interrupt signal")
		cancel()
	}()

	agent := NewAgent(cfg)
	agent.Run(ctx)
	lgr.Error("Agent stopped")
}
