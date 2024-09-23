package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nkiryanov/go-metrics/internal/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggerMiddleware(t *testing.T) {
	defer logger.Reset()

	// Define a dummy handler for testing
	dumbHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Define func to capture logged keys
	loggedKeys := func(entry observer.LoggedEntry) []string {
		keys := make([]string, 0, len(entry.Context))
		for _, field := range entry.Context {
			keys = append(keys, field.Key)
		}
		return keys
	}

	//
	core, recorded := observer.New(zapcore.InfoLevel)
	logger.Slog = zap.New(core).Sugar()

	lgrHandler := LoggerMiddleware(dumbHandler)

	t.Run("log http", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
		w := httptest.NewRecorder()

		lgrHandler.ServeHTTP(w, r)

		response := w.Result()
		defer response.Body.Close()
		require.Equal(t, http.StatusOK, response.StatusCode)
		require.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "got HTTP request", logEntry.Message)
		assert.ElementsMatch(t, []string{"method", "uri", "duration", "size", "status"}, loggedKeys(logEntry))

	})
}
