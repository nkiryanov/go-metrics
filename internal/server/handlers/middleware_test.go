package handlers

import (
	"bytes"
	"compress/gzip"
	"io"
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
	// Save global Slog and return back on exit
	prevSlog := logger.Slog
	defer func() { logger.Slog = prevSlog }()

	// Replace global logger with observed one. Reset to default when finish
	coreLogger, recorded := observer.New(zapcore.InfoLevel)
	logger.Slog = zap.New(coreLogger).Sugar()

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

	lgrHandler := LoggerMiddleware(dumbHandler)

	t.Run("log http", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
		w := httptest.NewRecorder()

		lgrHandler.ServeHTTP(w, r)

		response := w.Result()
		defer response.Body.Close() // nolint:errcheck
		require.Equal(t, http.StatusOK, response.StatusCode)
		require.Equal(t, 1, recorded.Len())
		logEntry := recorded.All()[0]
		assert.Equal(t, "got HTTP request", logEntry.Message)
		assert.ElementsMatch(t, []string{"method", "uri", "duration", "size", "status"}, loggedKeys(logEntry))

	})
}

func TestGzipMiddleWare(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	gzipCompress := func(data string) io.Reader {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		_, _ = w.Write([]byte(data))
		_ = w.Close()
		return &buf
	}

	gzipHandler := GzipMiddleware(okHandler)

	t.Run("compressed request ok", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/test-uri", gzipCompress(`{"some": "data"`))
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Accept-Encoding", "")
		w := httptest.NewRecorder()

		gzipHandler.ServeHTTP(w, r)

		response := w.Result()
		defer response.Body.Close() // nolint:errcheck

		body, err := io.ReadAll(response.Body)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "OK", string(body))
	})

	t.Run("send compressed request ok", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
		r.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		gzipHandler.ServeHTTP(w, r)

		response := w.Result()
		defer response.Body.Close() // nolint:errcheck

		require.Equal(t, "gzip", response.Header.Get("Content-Encoding"))

		// Read and decompress body
		gzipReader, err := gzip.NewReader(response.Body)
		require.NoError(t, err)
		defer gzipReader.Close() // nolint:errcheck

		body, err := io.ReadAll(gzipReader)
		require.NoError(t, err)

		require.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "OK", string(body))
	})
}
