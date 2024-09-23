package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestInitialize(t *testing.T) {
	t.Run("default level noop", func(t *testing.T) {
		assert.Equal(t, zapcore.InvalidLevel, Slog.Level())
	})

	t.Run("info level ok", func(t *testing.T) {
		defer Reset()

		_ = Initialize("info")

		assert.Equal(t, zapcore.InfoLevel, Slog.Level())
	})

	t.Run("error level ok", func(t *testing.T) {
		defer Reset()

		_ = Initialize("error")

		assert.Equal(t, zapcore.ErrorLevel, Slog.Level())
	})
}
