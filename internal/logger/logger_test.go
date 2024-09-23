package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestInitialize(t *testing.T) {
	t.Run("default level info", func(t *testing.T) {
		assert.Equal(t, zapcore.InfoLevel, Slog.Level())
	})

	t.Run("info level ok", func(t *testing.T) {
		globalSlog := Slog
		defer func() { Slog = globalSlog }()

		_ = Initialize("debug")

		assert.Equal(t, zapcore.DebugLevel, Slog.Level())
	})
}
