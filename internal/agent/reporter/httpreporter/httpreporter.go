package httpreporter

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
)

// Reporter error
// Store additional data to introspect error cause
type reportError struct {
	err     error // original error
	connErr bool  // wether HTTP connection error or something else
}

func (e *reportError) Error() string {
	return fmt.Sprintf("%v", e.err)
}

func (e *reportError) Unwrap() error {
	return e.err
}

func newReportErr(err error, connErr bool) *reportError {
	return &reportError{err, connErr}
}

// Metrics reporter to HTTP server
type HTTPReporter struct {
	reportAddr     string
	client         *http.Client
	maxRetries     int
	retryIntervals []time.Duration
}

func New(reportAddr string, client *http.Client, retryIntervals []time.Duration) *HTTPReporter {
	return &HTTPReporter{
		reportAddr:     reportAddr,
		client:         client,
		maxRetries:     len(retryIntervals),
		retryIntervals: retryIntervals,
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
	var body bytes.Buffer
	var err error

	gw := gzip.NewWriter(&body)

	encoder := json.NewEncoder(gw)
	err = encoder.Encode(data)
	if err != nil {
		logger.Slog.Errorw("error when marshaling json data", "error", err.Error(), "data", data)
		return newReportErr(err, false)
	}

	err = gw.Close()
	if err != nil {
		logger.Slog.Errorw("error when compressing data", "error", err.Error())
		return newReportErr(err, false)
	}

	request, err := http.NewRequest(http.MethodPost, reporter.reportAddr+url, &body)
	if err != nil {
		logger.Slog.Errorw("reporter: error when create request", "error", err.Error())
		return newReportErr(err, false)
	}

	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Type", "application/json")

	resp, err := reporter.client.Do(request)
	if err != nil {
		logger.Slog.Errorw("reporter: http client error", "error", err.Error())
		return newReportErr(err, true)
	}
	defer resp.Body.Close() // nolint:errcheck

	if status := resp.StatusCode; status != http.StatusOK {
		logger.Slog.Errorw("reporter: server responds with not OK", "status", status, "body", resp.Body)
		return newReportErr(err, false)
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
