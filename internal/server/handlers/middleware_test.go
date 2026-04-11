package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/ecdh"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/nkiryanov/go-metrics/internal/crypto"
	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/logger/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper handler that write "OK" to response
func okHandler(t *testing.T) http.HandlerFunc {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		require.NoError(t, err)
	})
}

func TestHandlers_LoggerMiddleware(t *testing.T) {
	logMsg := ""
	mockedLogger := &mocks.LoggerMock{
		InfoFunc: func(msg string, args ...any) { logMsg = msg },
	}

	lgrHandler := LoggerMiddleware(mockedLogger)(okHandler(t))

	t.Run("log http", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
		w := httptest.NewRecorder()

		lgrHandler.ServeHTTP(w, r)

		response := w.Result()
		defer response.Body.Close() // nolint:errcheck
		require.Equal(t, http.StatusOK, response.StatusCode)
		require.Equal(t, 1, len(mockedLogger.InfoCalls()))
		assert.Equal(t, "got HTTP request", logMsg)
	})
}

func TestHandlers_GzipMiddleWare(t *testing.T) {
	gzipCompress := func(data string) io.Reader {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		_, _ = w.Write([]byte(data))
		_ = w.Close()
		return &buf
	}

	gzipHandler := GzipMiddleware(okHandler(t))

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

func TestHandlers_HmacSHA256Middleware(t *testing.T) {
	t.Run("do nothing if empty secret key", func(t *testing.T) {
		secretKey := ""
		r := httptest.NewRequest(http.MethodPost, "/test-uri", strings.NewReader("hi!"))
		w := httptest.NewRecorder()

		handler := HmacSHA256Middleware(logger.NewNoOpLogger(), secretKey)(okHandler(t))
		handler.ServeHTTP(w, r)

		response := w.Result()
		defer response.Body.Close() // nolint:errcheck

		body, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "", response.Header.Get("HashSHA256"), "response should not has hmac header if secret key not set")
		assert.Equal(t, "OK", string(body))
	})

	tests := []struct {
		name                 string
		requestHmac          string
		expectedStatus       int
		expectedResponseHmac string
	}{
		{
			"error if hash not valid",
			"not-valid-hash-value",
			http.StatusBadRequest,
			"86dd45b12f91aed4f2e2c6b92ad113e02b3cbcf83dd210a5501d48ec4855ef34", // hmac of "message not authorized\n" and secret
		},
		{
			"ok if hash valid",
			"24aeae827da5c6a02e123468dd953cb706b2fae22ad1c1883d59810eafae6bc4", // hmac of "hi!" and secret
			http.StatusOK,
			"ffb8ab2cdd8a64b62d392d988408e0e52a68460c56bbcdec892c6b762ce4e340", // hmac of "OK" and secret
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			secretKey := "super-duper"
			r := httptest.NewRequest(http.MethodPost, "/test-upi", strings.NewReader("hi!"))
			r.Header.Set("HashSHA256", tc.requestHmac)
			w := httptest.NewRecorder()

			handler := HmacSHA256Middleware(logger.NewNoOpLogger(), secretKey)(okHandler(t))
			handler.ServeHTTP(w, r)

			response := w.Result()
			defer response.Body.Close() // nolint: errcheck

			assert.Equal(t, tc.expectedStatus, response.StatusCode)
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			assert.NotEmpty(t, string(body))
			require.Equal(t, tc.expectedResponseHmac, response.Header.Get("HashSHA256"), "all the responses must be signed")
		})
	}

	t.Run("do not fail if header omitted", func(t *testing.T) {
		secretKey := "super-duper"
		r := httptest.NewRequest(http.MethodPost, "/test-upi", strings.NewReader("hi!")) // No HashSHA256 header set
		w := httptest.NewRecorder()

		handler := HmacSHA256Middleware(logger.NewNoOpLogger(), secretKey)(okHandler(t)) // server running with HMAC support
		handler.ServeHTTP(w, r)

		response := w.Result()
		defer response.Body.Close() // nolint: errcheck

		assert.Equal(t, http.StatusOK, response.StatusCode)
		body, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		assert.Equal(t, "OK", string(body))
		assert.Equal(t, "ffb8ab2cdd8a64b62d392d988408e0e52a68460c56bbcdec892c6b762ce4e340", response.Header.Get("HashSHA256"), "Response has to be signed")
	})
}

func generateX25519Pair(t *testing.T) (*ecdh.PrivateKey, *ecdh.PublicKey) {
	t.Helper()
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	require.NoError(t, err)
	return priv, priv.PublicKey()
}

func TestHandlers_DecryptMiddleware(t *testing.T) {
	t.Run("noop when privKey is nil", func(t *testing.T) {
		original := []byte("plain body")
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(original))
		w := httptest.NewRecorder()

		var gotBody []byte
		handler := DecryptMiddleware(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, original, gotBody)
	})

	t.Run("decrypts body when privKey is set", func(t *testing.T) {
		priv, pub := generateX25519Pair(t)
		original := []byte("secret metrics payload")

		encrypted, err := crypto.Encrypt(pub, original)
		require.NoError(t, err)

		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(encrypted))
		w := httptest.NewRecorder()

		var gotBody []byte
		handler := DecryptMiddleware(priv)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, original, gotBody)
	})

	t.Run("returns 400 for garbage body", func(t *testing.T) {
		priv, _ := generateX25519Pair(t)
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("this is not encrypted"))
		w := httptest.NewRecorder()

		handler := DecryptMiddleware(priv)(okHandler(t))
		handler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
