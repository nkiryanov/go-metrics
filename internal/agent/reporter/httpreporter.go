package reporter

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
)

var gzPool = sync.Pool{
	New: func() interface{} { return gzip.NewWriter(io.Discard) },
}

type HTTPReporter struct {
	addr   string
	client *http.Client
}

func NewHTTPReporter(addr string, client *http.Client) *HTTPReporter {
	return &HTTPReporter{
		addr:   addr,
		client: client,
	}
}

// Sends a gzip encoded single metric update
// POST /{baseUrl}/update
// If the request encounters an error, it is returned.
func (rept *HTTPReporter) ReportOnce(m *models.Metric) error {
	body := &bytes.Buffer{}

	// Get gzip writer from them pool
	gz := gzPool.Get().(*gzip.Writer)
	gz.Reset(body)

	encoder := json.NewEncoder(gz)
	if err := encoder.Encode(m); err != nil {
		logger.Slog.Errorw("reporter: request error", "error", err.Error())
		return err
	}

	// Make sure body is written completely and return writer to pool
	gz.Close()
	gzPool.Put(gz)

	request, err := http.NewRequest(http.MethodPost, rept.addr+"/update", body)
	if err != nil {
		logger.Slog.Errorw("reporter: error when create request", "error", err.Error())
		return err
	}

	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Type", "application/json")

	resp, err := rept.client.Do(request)
	if err != nil {
		logger.Slog.Errorw("reporter: http client error", "error", err.Error())
		return err
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		logger.Slog.Errorw("reporter: server responds with not OK", "code", status, "body", resp.Body)
		return fmt.Errorf("reporter: metric update error = %s", resp.Body)
	}

	return nil
}

// ReportBatch sends concurrent metric update
// POST /{baseUrl}/update
// If errors while reporting occurred return one of them (randomly chosen)
func (rept *HTTPReporter) ReportBatch(metrics []models.Metric) error {
	body := &bytes.Buffer{}

	// Get gzip writer from them pool
	gz := gzPool.Get().(*gzip.Writer)
	gz.Reset(body)

	encoder := json.NewEncoder(gz)
	if err := encoder.Encode(metrics); err != nil {
		logger.Slog.Errorw("reporter: request error", "error", err.Error())
		return err
	}

	// Make sure body is written completely and return writer to pool
	gz.Close()
	gzPool.Put(gz)

	request, err := http.NewRequest(http.MethodPost, rept.addr+"/updates", body)
	if err != nil {
		logger.Slog.Errorw("reporter: error when create request", "error", err.Error())
		return err
	}

	request.Header.Set("Content-Encoding", "gzip")
	request.Header.Set("Content-Type", "application/json")

	resp, err := rept.client.Do(request)
	if err != nil {
		logger.Slog.Errorw("reporter: http client error", "error", err.Error())
		return err
	}
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		logger.Slog.Errorw("reporter: server responds with not OK", "code", status, "body", resp.Body)
		return fmt.Errorf("reporter: metric update error = %s", resp.Body)
	}

	return nil
}
