package reporter

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/nkiryanov/go-metrics/internal/logger"

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
// POST /{baseUrl}/update/{metricType}/{metricName}/{metricValue}
// If the request encounters an error, it is returned.
func (rept *HTTPReporter) ReportOnce(m *Metric) error {
	resp, err := rept.client.R().
		SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"mType":  m.Type,
			"mName":  m.Name,
			"mValue": m.Value.String(),
		}).
		Post(fmt.Sprintf("%s/update/{mType}/{mName}/{mValue}", rept.addr))

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
// POST /{baseUrl}/update/{metricType}/{metricName}/{metricValue}
// If errors while reporting occurred return one of them (randomly chosen)
func (rept *HTTPReporter) ReportBatch(ms []*Metric) error {
	var wg sync.WaitGroup
	var once sync.Once
	var err error

	onceBody := func(e error) { err = e }

	for _, m := range ms {
		wg.Add(1)
		go func(m *Metric) {
			defer wg.Done()

			if err := rept.ReportOnce(m); err != nil {
				once.Do(func() {
					onceBody(err)
				})
			}
		}(m)
	}

	wg.Wait()

	if err == nil {
		logger.Slog.Info("reporter: metrics updated", "count", len(ms))
	}

	return err
}
