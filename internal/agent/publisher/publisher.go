package publisher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/nkiryanov/go-metrics/internal/storage"
)

type Publisher interface {
	Run(ctx context.Context) error
}

var ErrPublisherStopped = errors.New("publisher: Publisher stopped")

type HttpPublisher struct {
	pubAddr     string
	pubInterval time.Duration

	storage storage.Storage
	client  *http.Client
}

func NewHttpPublisher(pubAddr string, pubInterval time.Duration, storage storage.Storage) (*HttpPublisher, error) {
	pubUrl, err := url.Parse(pubAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse publisher address: %w", err)
	}
	if pubUrl.Scheme == "" || pubUrl.Host == "" {
		return nil, errors.New("publisher: Invalid publisher address")
	}

	return &HttpPublisher{
		pubAddr:     pubUrl.String(),
		pubInterval: pubInterval,
		storage:     storage,
		client:      &http.Client{},
	}, nil
}

func (p HttpPublisher) postMetric(mType storage.MetricType, name storage.MetricName) (status int, err error) {
	var value string

	switch mType {
	case storage.CounterTypeName:
		value = func() string { counter, _ := p.storage.GetCounter(name); return counter.String() }()
	case storage.GaugeTypeName:
		value = func() string { gauge, _ := p.storage.GetGauge(name); return gauge.String() }()
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/update/%s/%s/%s", p.pubAddr, mType, name, value), nil)
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

func (p HttpPublisher) batchPublish() {
	var wg sync.WaitGroup

	p.storage.IterateGauges(func(name storage.MetricName, value storage.Gaugeable) {
		wg.Add(1)

		go func(name storage.MetricName) {
			defer wg.Done()

			status, err := p.postMetric(storage.GaugeTypeName, name)
			if err != nil || status != http.StatusOK {
				slog.Error("Failed to post gauge", "name", name, "error", err, "status", status)
				return
			}

			slog.Info("Gauge posted", "name", name, "status", status)
		}(name)
	})

	p.storage.IterateCounters(func(name storage.MetricName, value storage.Countable) {
		wg.Add(1)

		go func(name storage.MetricName) {
			defer wg.Done()

			status, err := p.postMetric(storage.CounterTypeName, name)
			if err != nil || status != http.StatusOK {
				slog.Error("Failed to post counter", "name", name, "error", err, "status", status)
				return
			}

			slog.Info("Counter posted", "name", name, "status", status)
		}(name)
	})

	wg.Wait()
}

func (p HttpPublisher) Run(ctx context.Context) error {
	go func() {
		ticker := time.NewTicker(p.pubInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p.batchPublish()
				slog.Info("Metrics published")
				ticker.Reset(p.pubInterval)
			}
		}
	}()

	<-ctx.Done()

	return ErrPublisherStopped
}
