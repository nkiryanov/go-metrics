package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log, Slog = Reset()

func Reset() (*zap.Logger, *zap.SugaredLogger) {
	// Reset logger to default values.
	plain := zap.NewNop()
	sugar := plain.Sugar()
	return plain, sugar
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

	Log = lgr
	Slog = lgr.Sugar()
	return nil
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		Log.Info(
			"got incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Duration("duration", time.Since(start)),
		)
	})
}
