package httpreporter

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
)

// context to send over post request. It may changed with middleware
type postContext struct {
	headers map[string]string // headers to set to the request
	data    any               // initial data to send over post request. It must be read and converted somehow to body
	body    *bytes.Buffer     // body to send over post request
}

// Reporter error
// Store additional data to introspect error cause
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
	reportAddr     string
	client         *http.Client
	maxRetries     int
	retryIntervals []time.Duration
	secretKey      string
}

func New(reportAddr string, client *http.Client, retryIntervals []time.Duration, secretKey string) *HTTPReporter {
	return &HTTPReporter{
		reportAddr:     reportAddr,
		client:         client,
		maxRetries:     len(retryIntervals),
		retryIntervals: retryIntervals,
		secretKey:      secretKey,
	}
}

// POST /{baseUrl}/update
func (reporter *HTTPReporter) ReportOnce(m models.Metric) error {
	return reporter.postWithRetry("/update", m)
}

// POST /{baseUrl}/update
func (reporter *HTTPReporter) ReportBatch(metrics []models.Metric) error {
	return reporter.postWithRetry("/updates", metrics)
}

// POSTs data to url server as gzipped JSON
func (reporter *HTTPReporter) post(url string, data any) error {
	postContext := postContext{
		headers: make(map[string]string),
		data:    data,
		body:    nil,
	}

	err := jsonGzipMiddleware(&postContext)
	if err != nil {
		return newReportError(err, false)
	}

	request, err := http.NewRequest(http.MethodPost, reporter.reportAddr+url, postContext.body)
	if err != nil {
		return newReportError(err, false)
	}

	for key, value := range postContext.headers {
		request.Header.Set(key, value)
	}

	resp, err := reporter.client.Do(request)
	if err != nil {
		return newReportError(err, true)
	}
	defer resp.Body.Close() // nolint:errcheck

	if status := resp.StatusCode; status != http.StatusOK {
		return newReportError(err, false)
	}

	return nil
}

// POSTs data to url server as gzipped JSON and retry if connection error occurs
func (reporter *HTTPReporter) postWithRetry(url string, data any) error {
	var err error

	for attempt := 0; attempt <= reporter.maxRetries; attempt++ {
		if attempt > 0 {
			delay := reporter.retryIntervals[attempt-1]
			logger.Slog.Infow("retrying request", "attempt", attempt, "delay", delay)
			time.Sleep(delay)
		}

		err = reporter.post(url, data)

		// Return if no error occurred or it's not connection error
		var errReport *reportError
		if err == nil || (errors.As(err, &errReport) && !errReport.connErr) {
			return err
		}

		logger.Slog.Warnw("request connection error, retrying", "error", err, "attempt", attempt+1)
	}

	logger.Slog.Errorw("all retry attempts failed", "error", err, "attempts", reporter.maxRetries+1)
	return err
}
