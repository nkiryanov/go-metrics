package reporter

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
)

type HTTPReporter struct {
	reportAddr string
	client     *http.Client
}

func NewHTTPReporter(reportAddr string, client *http.Client) *HTTPReporter {
	return &HTTPReporter{
		reportAddr: reportAddr,
		client:     client,
	}
}

// POST /{baseUrl}/update
func (reporter *HTTPReporter) ReportOnce(m models.Metric) error {
	return reporter.postRequest("/update", m)
}

// POST /{baseUrl}/update
func (reporter *HTTPReporter) ReportBatch(metrics []models.Metric) error {
	return reporter.postRequest("/updates", metrics)
}

// POSTs data to url server as gzipped JSON
func (reporter *HTTPReporter) postRequest(url string, data any) error {
	var body bytes.Buffer
	var err error

	gw := gzip.NewWriter(&body)

	encoder := json.NewEncoder(gw)
	err = encoder.Encode(data)
	if err != nil {
		logger.Slog.Errorw("error when marshaling json data", "error", err.Error(), "data", data)
		return err
	}

	err = gw.Close()
	if err != nil {
		logger.Slog.Errorw("error when compressing data", "error", err.Error())
		return err
	}

	request, err := http.NewRequest(http.MethodPost, reporter.reportAddr+url, &body)
	if err != nil {
		logger.Slog.Errorw("reporter: error when create request", "error", err.Error())
		return err
	}

	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Type", "application/json")

	resp, err := reporter.client.Do(request)
	if err != nil {
		logger.Slog.Errorw("reporter: http client error", "error", err.Error())
		return err
	}
	defer resp.Body.Close() // nolint:errcheck

	if status := resp.StatusCode; status != http.StatusOK {
		logger.Slog.Errorw("reporter: server responds with not OK", "status", status, "body", resp.Body)
		return fmt.Errorf("reporter: metric update error = %s", resp.Body)
	}

	return nil
}
