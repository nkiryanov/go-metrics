package handlers

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func gzipData(t *testing.B, data string) []byte {
	// Some data easy to compress
	var buf bytes.Buffer

	gw := gzip.NewWriter(&buf)

	_, err := gw.Write([]byte(data))
	require.NoError(t, err)

	err = gw.Close()
	require.NoError(t, err)

	return buf.Bytes()
}

func BenchmarkGzipMiddleware(b *testing.B) {
	b.ReportAllocs()

	// Some data that easy to compress
	data := strings.Repeat("easy-to-compress", 1000)
	gzippedData := gzipData(b, data)

	// Simple handler that return json wrapped with gzip middleware
	handler := GzipMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(data))
	}))

	for b.Loop() {
		req := httptest.NewRequest(http.MethodGet, "/update", bytes.NewReader(gzippedData))
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
	}
}
