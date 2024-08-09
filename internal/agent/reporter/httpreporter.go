package reporter

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"
)

type HTTPReporter struct {
	addr string
	client *resty.Client
}

func NewHTTPReporter(addr string, client *resty.Client) *HTTPReporter {
	return &HTTPReporter{ 
		addr: 	addr,
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
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("reporter: metric update error = %s", resp.Body())
	}

	return nil
}

// ReportBatch sends concurrent metric update
// POST /{baseUrl}/update/{metricType}/{metricName}/{metricValue}
// All errors captured and returned.
func (rept *HTTPReporter) ReportBatch(ms []*Metric) []error {
	var wg sync.WaitGroup

	errs := make([]error, 0)

	ewriter := func() func(error) {
		var mu sync.Mutex
		return func(err error) {
			mu.Lock()
			errs = append(errs, err)
			mu.Unlock()
		}
	}()

	for _, m := range ms {
		wg.Add(1)
		go func(m *Metric) {
			defer wg.Done()

			if err := rept.ReportOnce(m); err != nil {
				ewriter(err)
			}
		}(m)
	}

	wg.Wait()
	return errs
}
