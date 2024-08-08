package reporter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/nkiryanov/go-metrics/internal/storage"
)

type Reporter interface {
	Run(ctx context.Context) error
}

var ErrReporterStopped = errors.New("reporter: Reporter stopped")

type HTTPReporter struct {
	repAddr     string
	repInterval time.Duration

	storage storage.Storage
	client  *http.Client
}

func NewHTTPReporter(repAddr string, repInterval time.Duration, storage storage.Storage) (*HTTPReporter, error) {
	return &HTTPReporter{
		repAddr:     repAddr,
		repInterval: repInterval,
		storage:     storage,
		client:      &http.Client{},
	}, nil
}

func (p HTTPReporter) reportMetric(mType string, name string) (status int, err error) {
	var value string

	switch mType {
	case storage.CounterTypeName:
		value = func() string { counter, _ := p.storage.GetCounter(name); return counter.String() }()
	case storage.GaugeTypeName:
		value = func() string { gauge, _ := p.storage.GetGauge(name); return gauge.String() }()
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/update/%s/%s/%s", p.repAddr, mType, name, value), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "text/plain")

	response, err := p.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	code := response.StatusCode
	if _, err = io.Copy(io.Discard, response.Body); err != nil {
		return code, err
	}

	return code, nil
}

func (p HTTPReporter) batchReport() {
	var wg sync.WaitGroup

	p.storage.IterateGauges(func(name string, value storage.Gauge) {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			status, err := p.reportMetric(storage.GaugeTypeName, name)
			if err != nil || status != http.StatusOK {
				slog.Error("Failed to post gauge", "name", name, "error", err, "status", status)
				return
			}

			slog.Info("Gauge posted", "name", name, "status", status)
		}(name)
	})

	p.storage.IterateCounters(func(name string, value storage.Counter) {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			status, err := p.reportMetric(storage.CounterTypeName, name)
			if err != nil || status != http.StatusOK {
				slog.Error("Failed to post counter", "name", name, "error", err, "status", status)
				return
			}

			slog.Info("Counter posted", "name", name, "status", status)
		}(name)
	})

	wg.Wait()
}

func (p HTTPReporter) Run(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(p.repInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.batchReport()
				slog.Info("Metrics published")
				ticker.Reset(p.repInterval)
			}
		}
	}()

	<-ctx.Done()

	return ErrReporterStopped
}
