package reporter

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-resty/resty/v2"
)

type HTTPReporter struct {
	client *resty.Client
}

func NewHTTPReporter(addr string) *HTTPReporter {
	return &HTTPReporter{
		client: resty.New().SetBaseURL(addr),
	}
}

func (rept *HTTPReporter) ReportOnce(m *Metric) error {
	resp, err := rept.client.R().
		SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"mType":  m.Type,
			"mName":  m.Name,
			"mValue": m.Value.String(),
		}).
		Post("/update/{mType}/{mName}/{mValue}")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("reporter: metric update error = %s", resp.Body())
	}

	return nil
}

func (rept *HTTPReporter) ReportBatch(ms []*Metric) []error {
	var wg sync.WaitGroup

	var errs []error = make([]error, 0)

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
