package memstorage

import (
	"context"
	"errors"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/server/storage"
)

// MemStorage implements storage interface with in-memory metrics and optional file persistence
type MemStorage struct {
	lgr logger.Logger
	// Metric storages
	counters map[string]int64
	gauges   map[string]float64
	mu       sync.RWMutex

	// File persistence
	// Save interval means how often metrics should be saved (0 — should saved synchronously, on each update)
	file         *os.File
	fileLock     sync.Mutex
	saveInterval time.Duration
	stopCh       chan struct{}
}

// New creates a MemStorage instance, optionally with file persistence.
// If filename is provided, metrics will be saved at specified saveInterval (0 = synchronous saves)
// If filename, saveInterval, restore is empty then create in memory only storage
func New(filename string, saveInterval time.Duration, restore bool, lgr logger.Logger) (*MemStorage, error) {
	if filename == "" && (saveInterval != 0 || restore) {
		return nil, errors.New("can't create in-memory only storage, cause one New args not empty")
	}

	// Open file as persistent storage for metrics
	var file *os.File
	if filename != "" {
		var err error
		file, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return nil, err
		}
	}

	s := &MemStorage{
		counters:     make(map[string]int64),
		gauges:       make(map[string]float64),
		file:         file,
		saveInterval: saveInterval,
		stopCh:       make(chan struct{}),
		lgr:          lgr,
	}

	if restore {
		err := s.loadFromFile()
		if err != nil {
			return nil, err
		}
	}

	if saveInterval > 0 {
		go s.backgroundSaver()
	}

	return s, nil
}

func (s *MemStorage) Close() error {
	if s.stopCh != nil {
		close(s.stopCh)
	}

	if s.file != nil {
		return s.saveToFile()
	}

	return nil
}

// Memory storage is ready for use just after initialization
// No errors is possible
func (s *MemStorage) Ping(_ context.Context) error {
	return nil
}

// CountMetric returns a total count of all metrics in storage
func (s *MemStorage) CountMetric(_ context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.counters) + len(s.gauges), nil
}

// GetMetric gets a metric by its type and name. Returns error if metric not found
func (s *MemStorage) GetMetric(_ context.Context, mType string, mName string) (models.Metric, error) {
	var err = storage.ErrNoMetric
	metric := models.Metric{Type: mType, Name: mName}

	s.mu.RLock()
	defer s.mu.RUnlock()
	var readOk bool

	switch mType {
	case models.CounterTypeName:
		metric.Delta, readOk = s.counters[mName]
		if readOk {
			return metric, nil
		}
	case models.GaugeTypeName:
		metric.Value, readOk = s.gauges[mName]
		if readOk {
			return metric, nil
		}
	}

	return metric, err
}

// Save valid metric in storage and return updated value
func (s *MemStorage) UpdateMetric(_ context.Context, in models.Metric) (models.Metric, error) {
	var metric models.Metric

	err := in.Validate()
	if err != nil {
		return metric, err
	}

	metric = s.update([]models.Metric{in})[0]

	// is saveInterval = 0, than save metrics in synchronously
	if s.saveInterval == 0 {
		err := s.saveToFile()
		if err != nil {
			s.lgr.Error("metrics saving failed", "error", err.Error())
			return metric, err
		}
	}

	return metric, nil
}

// ListMetric returns all metrics in storage sorted by name
func (s *MemStorage) ListMetric(_ context.Context) ([]models.Metric, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := make([]models.Metric, 0, len(s.counters)+len(s.gauges))

	for name, counter := range s.counters {
		metrics = append(metrics, models.Metric{Name: name, Type: models.CounterTypeName, Delta: counter})
	}

	for name, gauge := range s.gauges {
		metrics = append(metrics, models.Metric{Name: name, Type: models.GaugeTypeName, Value: gauge})
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})

	return metrics, nil
}

// UpdateMetricBulk takes a slice of metrics, validates them, updates them in storage,
// and returns the updated metrics. If saveInterval is 0, saves to file synchronously.
func (s *MemStorage) UpdateMetricBulk(_ context.Context, metrics []models.Metric) ([]models.Metric, error) {
	var err error

	// Validate all the metrics
	for _, m := range metrics {
		err = m.Validate()
		if err != nil {
			return metrics, err
		}
	}

	updated := s.update(metrics)

	// is saveInterval = 0, than save metrics in synchronously
	if s.saveInterval == 0 {
		err = s.saveToFile()
		if err != nil {
			s.lgr.Error("metrics saving failed", "error", err.Error())
			return updated, err
		}
	}

	return updated, nil
}

// Update valid metric in storage (in memory) and return updated value
func (s *MemStorage) update(validated []models.Metric) []models.Metric {
	updated := make([]models.Metric, len(validated))

	s.mu.Lock()
	defer s.mu.Unlock()

	for idx, m := range validated {
		updated[idx].Type = m.Type
		updated[idx].Name = m.Name

		switch m.Type {
		case models.CounterTypeName:
			s.counters[m.Name] += m.Delta
			updated[idx].Delta = s.counters[m.Name]
		case models.GaugeTypeName:
			s.gauges[m.Name] = m.Value
			updated[idx].Value = s.gauges[m.Name]
		}
	}

	return updated
}
