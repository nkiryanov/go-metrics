package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
)

type loggedData struct {
	responseStatus int
	responseSize   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	loggedData loggedData
}

func (w *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.loggedData.responseSize += size
	return size, err
}

func (w *loggingResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.loggedData.responseStatus = statusCode
}

func LoggerMiddleware(lgr logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lw := loggingResponseWriter{
				ResponseWriter: w,
				loggedData:     loggedData{responseStatus: 200, responseSize: 0},
			}

			next.ServeHTTP(&lw, r)

			lgr.Info(
				"got HTTP request",
				"method", r.Method,
				"uri", r.RequestURI,
				"duration", time.Since(start),
				"status", lw.loggedData.responseStatus,
				"size", lw.loggedData.responseSize,
			)
		})
	}
}

type compressWriter struct {
	w  http.ResponseWriter
	gw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		gw: gzip.NewWriter(w),
	}
}

func (cw *compressWriter) Header() http.Header {
	return cw.w.Header()
}

func (cw *compressWriter) Write(b []byte) (int, error) {
	return cw.gw.Write(b)
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	cw.w.Header().Set("Content-Encoding", "gzip")
	cw.w.WriteHeader(statusCode)
}

func (cw *compressWriter) Close() error {
	return cw.gw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	gr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		gr: gr,
	}, nil
}

func (cr *compressReader) Read(b []byte) (int, error) {
	return cr.gr.Read(b)
}

func (cr *compressReader) Close() error {
	if err := cr.r.Close(); err != nil {
		_ = cr.gr.Close()
		return err
	}

	return cr.gr.Close()
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

// Capture Write calls and store content to internal 'body' buffer: it is required to calculate HMAC value
// Capture WriteHeaders calls and store at internal 'statusCode'. It has to be done to not write send headers before HMAC calculated and added to Headers
// Release() method calculate HMAC for the entire body and call WriteHeader and Write methods accordingly
type hmacWriter struct {
	w          http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	lgr        logger.Logger
}

func (hw *hmacWriter) Header() http.Header {
	return hw.w.Header()
}

// Capture all the data to internal buffer
func (hw *hmacWriter) Write(p []byte) (int, error) {
	return hw.body.Write(p)
}

// Capture status code to internal buffer
func (hw *hmacWriter) WriteHeader(statusCode int) {
	hw.statusCode = statusCode
}

func (hw *hmacWriter) Release(secretKey []byte) {
	// Calculate HMAC header
	h := hmac.New(sha256.New, secretKey)
	_, err := h.Write(hw.body.Bytes())
	if err != nil {
		hw.lgr.Error("HMAC message signing failed", "error", err)
	} else {
		hw.Header().Set("HashSHA256", hex.EncodeToString(h.Sum(nil)))
	}

	// Write captured data to internal writer
	hw.w.WriteHeader(hw.statusCode)
	_, err = hw.w.Write(hw.body.Bytes())

	if err != nil {
		hw.lgr.Error("write response error", "error", err.Error())
		http.Error(hw.w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HmacSHA256Middleware(lgr logger.Logger, secretKey string) func(http.Handler) http.Handler {
	if secretKey == "" {
		// Noop middleware: just call next middleware
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}
	}

	key := []byte(secretKey)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// All responses has to be signed. Replace writer with HMAC Writer
			hmacw := &hmacWriter{w: w, body: &bytes.Buffer{}, lgr: lgr}
			defer hmacw.Release(key)

			// If HashSHA256 is set then HMAC for request must be calculated and verified
			if expectedMac := r.Header.Get("HashSHA256"); expectedMac != "" {
				// Read entire body to buffer for HMAC calculation
				// Also replace request's body with copy of the read body
				buf, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(hmacw, "invalid body", http.StatusBadRequest)
					return
				}
				_ = r.Body.Close()
				r.Body = io.NopCloser(bytes.NewReader(buf))

				// Verify HMAC signature
				h := hmac.New(sha256.New, key)
				_, err = h.Write(buf)
				if err != nil {
					http.Error(hmacw, err.Error(), http.StatusInternalServerError)
				}
				actualMac := hex.EncodeToString(h.Sum(nil))
				if !hmac.Equal([]byte(expectedMac), []byte(actualMac)) {
					http.Error(hmacw, "message not authorized", http.StatusBadRequest)
					return
				}
			}

			next.ServeHTTP(hmacw, r)
		})
	}
}
