package handlers

import (
	"net/http"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData responseData
	}
)

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
