package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestInitialize(t *testing.T) {
	t.Run("default level noop", func(t *testing.T) {
		assert.Equal(t, zapcore.InvalidLevel, Log.Level())
		assert.Equal(t, zapcore.InvalidLevel, Slog.Level())
	})

	t.Run("info level ok", func(t *testing.T) {
		defer Reset()

		Initialize("info")

		assert.Equal(t, zapcore.InfoLevel, Log.Level())
		assert.Equal(t, zapcore.InfoLevel, Slog.Level())
	})

	t.Run("error level ok", func(t *testing.T) {
		defer Reset()

		Initialize("error")

		assert.Equal(t, zapcore.ErrorLevel, Log.Level())
		assert.Equal(t, zapcore.ErrorLevel, Slog.Level())
	})
}

func TestRequestLogger(t *testing.T) {
	defer Reset()

	// Define a dummy handler for testing
	dumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Define func to capture logged keys
	loggedKeys := func(entry observer.LoggedEntry) []string {
		keys := make([]string, 0, len(entry.Context))
		for _, field := range entry.Context {
			keys = append(keys, field.Key)
		}
		return keys
	}

	// Prepare logger
	core, recorded := observer.New(zapcore.InfoLevel)
	Log = zap.New(core)

	lgrHandler := RequestLogger(dumbHandler)

	t.Run("log http", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
		w := httptest.NewRecorder()

		lgrHandler.ServeHTTP(w, r)

		response := w.Result()
		require.Equal(t, http.StatusOK, response.StatusCode)
		require.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "got incoming HTTP request", logEntry.Message)
		assert.ElementsMatch(t, []string{"method", "uri", "duration"}, loggedKeys(logEntry))
	})
}
