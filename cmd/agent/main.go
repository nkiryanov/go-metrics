package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/config"
	"github.com/nkiryanov/go-metrics/internal/logger"
)

// Should be set on a build stage[
var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

const (
	LogLevel = "info"

	ReportAddr     = "http://localhost:8080"
	ReportInterval = 10 * time.Second
	RateRateLimit  = 50
	SecretKey      = ""

	CollectInterval = 2 * time.Second
)

func main() {
	cfg := &config.Config{
		LogLevel:        LogLevel,
		ReportAddr:      ReportAddr,
		ReportInterval:  ReportInterval,
		ReportRateLimit: RateRateLimit,
		SecretKey:       SecretKey,
		CollectInterval: CollectInterval,
	}
	cfg.MustLoad()

	lgr := logger.NewLogger(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, cfg, lgr); err != nil {
		lgr.Error("Agent stopped with error", "error", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config, lgr logger.Logger) error {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	agent := NewAgent(cfg)

	if err := agent.Run(ctx); err != nil && err != ErrAgentStopped {
		return err
	}

	lgr.Info("Agent stopped gracefully")
	return nil
}
