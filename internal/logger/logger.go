package logger

import (
	"go.uber.org/zap"
	"log"
)

// Initialize global Slog with default production logger
var Slog = func() *zap.SugaredLogger {
	lgr, err := zap.NewProduction()
	if err != nil {
		log.Fatal("global logger initialization failed", err)
	}
	return lgr.Sugar()
}()

// Initialize Slog with given level
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	lgr, err := cfg.Build()
	if err != nil {
		return err
	}

	Slog = lgr.Sugar()
	return nil
}
