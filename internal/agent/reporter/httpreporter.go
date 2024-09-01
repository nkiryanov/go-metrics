package reporter

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"

	"github.com/go-resty/resty/v2"
)

type HTTPReporter struct {
	addr   string
	client *resty.Client
}

func NewHTTPReporter(addr string, client *resty.Client) *HTTPReporter {
	return &HTTPReporter{
		addr:   addr,
		client: client,
	}
}

// Sends a single metric update
// POST /{baseUrl}/update/{metricType}/{metricID}/{metricValue}
// If the request encounters an error, it is returned.
func (rept *HTTPReporter) ReportOnce(m *models.Metric) error {
	resp, err := rept.client.R().
		SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"mType":  m.MType,
			"mID":    m.ID,
			"mValue": m.String(),
		}).
		Post(fmt.Sprintf("%s/update/{mType}/{mID}/{mValue}", rept.addr))

	if err != nil {
		logger.Slog.Error("reporter: http client error", "error", err.Error())
		return err
	}

	if status := resp.StatusCode(); status != http.StatusOK {
		logger.Slog.Error("reporter: server responds with not OK", "code", status, "body", resp.Body())
		return fmt.Errorf("reporter: metric update error = %s", resp.Body())
	}

	return nil
}

// ReportBatch sends concurrent metric update
// POST /{baseUrl}/update/{metricType}/{metricID}/{metricValue}
// If errors while reporting occurred return one of them (randomly chosen)
func (rept *HTTPReporter) ReportBatch(metrics []models.Metric) error {
	var wg sync.WaitGroup
	var once sync.Once
	var err error

	onceBody := func(e error) { err = e }

	for _, metric := range metrics {
		wg.Add(1)
		go func(m *models.Metric) {
			defer wg.Done()

			if err := rept.ReportOnce(m); err != nil {
				once.Do(func() {
					onceBody(err)
				})
			}
		}(&metric)
	}

	wg.Wait()

	if err == nil {
		logger.Slog.Info("reporter: metrics updated", "count", len(metrics))
	}

	return err
}
