package handlers

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
)

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData{status: 200, size: 0},
		}

		next.ServeHTTP(&lw, r)

		logger.Slog.Infow(
			"got HTTP request",
			"method", r.Method,
			"uri", r.RequestURI,
			"duration", time.Since(start),
			"status", lw.responseData.status,
			"size", lw.responseData.size,
		)
	})
}

type compressWriter struct {
	w  http.ResponseWriter
	cw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		cw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(b []byte) (int, error) {
	return c.cw.Write(b)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.Header().Set("Content-Encoding", "gzip")
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.cw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	cr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	cr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		cr: cr,
	}, nil
}

func (c *compressReader) Read(b []byte) (int, error) {
	return c.cr.Read(b)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		_ = c.cr.Close()
		return err
	}

	return c.cr.Close()
}

// Naive implementation. Prefer chi.Compressor middleware instead
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If request gzip encoded replace r.Reader with compressReader
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer cr.Close() // nolint:errcheck
		}

		// If client accept gzip encoded responses than compress response
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(w)
			defer cw.Close() // nolint:errcheck

			w = cw
		}

		next.ServeHTTP(w, r)
	})
}
