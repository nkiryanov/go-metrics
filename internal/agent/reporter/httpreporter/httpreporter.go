package httpreporter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
)

// Context for sending post request may be changed by middleware
type postContext struct {
	headers map[string]string // headers to set on the request
	data    any               // initial data to send over post request must be converted to body
	body    *bytes.Buffer     // body to send over post request
}

// Reporter error with additional data to introspect error cause
type reportError struct {
	err     error // original error
	connErr bool  // whether it's an HTTP connection error or something else
}

func (e *reportError) Error() string {
	return fmt.Sprintf("%v", e.err)
}

func (e *reportError) Unwrap() error {
	return e.err
}

func newReportError(err error, connErr bool) *reportError {
	return &reportError{err, connErr}
}

// Metrics reporter to HTTP server
type HTTPReporter struct {
	reportAddr     string          // address to report to
	maxRetries     int             // maximum number of retries on connection error
	retryIntervals []time.Duration // intervals to retry on connection error
	sem            *Semaphore      // limits concurrent requests to server
	secretKey      string          // secret key to sign requests with hmac sha256

	client *http.Client
	lgr    logger.Logger
}

func New(
	reportAddr string,
	retryIntervals []time.Duration,
	rateLimit int,
	secretKey string,
	client *http.Client,
	lgr logger.Logger,
) *HTTPReporter {
	return &HTTPReporter{
		reportAddr:     reportAddr,
		maxRetries:     len(retryIntervals),
		retryIntervals: retryIntervals,
		sem:            NewSemaphore(rateLimit),
		secretKey:      secretKey,
		client:         client,
		lgr:            lgr,
	}
}

// POST /{baseUrl}/update
func (r *HTTPReporter) ReportOnce(metric models.Metric) error {
	return r.postWithRetry("/update", metric)
}

// POST /{baseUrl}/updates
func (r *HTTPReporter) ReportBatch(metrics []models.Metric) error {
	return r.postWithRetry("/updates", metrics)
}

// POST data to url server as gzipped json
func (r *HTTPReporter) post(url string, data any) error {
	r.sem.Acquire()
	defer r.sem.Release()

	postContext := postContext{
		headers: make(map[string]string),
		data:    data,
		body:    nil,
	}

	err := r.jsonGzipMiddleware(&postContext)
	if err != nil {
		return newReportError(err, false)
	}
	err = r.hmacSha256Middleware(&postContext)
	if err != nil {
		return newReportError(err, false)
	}

	request, err := http.NewRequest(http.MethodPost, r.reportAddr+url, postContext.body)
	if err != nil {
		return newReportError(err, false)
	}

	for key, value := range postContext.headers {
		request.Header.Set(key, value)
	}

	resp, err := r.client.Do(request)
	if err != nil {
		return newReportError(err, true)
	}
	defer resp.Body.Close() // nolint:errcheck

	if status := resp.StatusCode; status != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return newReportError(fmt.Errorf("http %d: %s", resp.StatusCode, string(body)), false)
	}

	return nil
}

// POST data to url server as gzipped json and retry if connection error occurs
func (r *HTTPReporter) postWithRetry(url string, data any) error {
	var err error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			delay := r.retryIntervals[attempt-1]
			r.lgr.Info("retrying request", "attempt", attempt, "delay", delay)
			time.Sleep(delay)
		}

		err = r.post(url, data)

		// Return if no error occurred or it's not connection error
		var errReport *reportError
		if err == nil || (errors.As(err, &errReport) && !errReport.connErr) {
			return err
		}

		r.lgr.Warn("request connection error, retrying", "error", err, "attempt", attempt+1)
	}

	r.lgr.Error("all retry attempts failed", "error", err, "attempts", r.maxRetries+1)
	return err
}
