package logger

import (
	"go.uber.org/zap"
)

var Slog = Reset()

func Reset() *zap.SugaredLogger {
	// Reset logger to default values.
	return zap.NewNop().Sugar()
}

func Initialize(level string) error {
	// Initialize Log (zap.Logger) and Slog (zap.SugaredLogger) with the given log level
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
