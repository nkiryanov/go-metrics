package memstorage

import (
	"context"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

type counterStore struct {
	lock    sync.RWMutex
	storage map[string]int64
}

type gaugeStore struct {
	lock    sync.RWMutex
	storage map[string]float64
}

type saveFunc func(s *MemStorage) error
type restoreFunc func(s *MemStorage) error

type MemStorage struct {
	// Metric storages
	gauges   gaugeStore
	counters counterStore

	// File as a persistent storage
	// Save interval means how often metrics should be saved (0 — should saved synchronously, on each update)
	file         *os.File
	fileLock     sync.Mutex
	saveInterval time.Duration
	saver        saveFunc
	restorer     restoreFunc

	// MemStorage context. Should be cancelled when Close() is called
	ctx    context.Context
	cancel context.CancelFunc
}

func New(filePath string, interval time.Duration, restore bool) (*MemStorage, error) {
	var err error
	var file *os.File

	// Open file as persistent storage for metrics
	file, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	// Initialize context with cancel function
	ctx, cancel := context.WithCancel(context.Background())

	storage := &MemStorage{
		counters: counterStore{storage: make(map[string]int64)},
		gauges:   gaugeStore{storage: make(map[string]float64)},

		file:         file,
		saveInterval: interval,
		saver:        memSave,
		restorer:     memRestore,

		ctx:    ctx,
		cancel: cancel,
	}

	// Restore storage data from file if needed
	if restore {
		if err = storage.restore(); err != nil {
			return nil, err
		}
		logger.Slog.Info("storage: restored from file")
	}

	// Run interval saver if needed
	if storage.saveInterval > 0 {
		go func(s *MemStorage) {
			ticker := time.NewTicker(s.saveInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					if err := s.save(); err != nil {
						logger.Slog.Errorw("storage: save failed", "error", err.Error())
					} else {
						logger.Slog.Debug("storage: saved")
					}
				case <-s.ctx.Done():
					return
				}
			}
		}(storage)
		logger.Slog.Infow("storage: interval saver started", "interval", storage.saveInterval.String())
	}

	return storage, nil
}

func (s *MemStorage) save() error {
	return s.saver(s)
}

func (s *MemStorage) restore() error {
	return s.restorer(s)
}

func (s *MemStorage) Close() error {
	s.cancel()
	if err := s.save(); err != nil {
		return err
	}
	return s.file.Close()
}

// Memory storage is ready for use just after initialization
// No errors is possible
func (s *MemStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *MemStorage) CountMetric(ctx context.Context) (int, error) {
	s.counters.lock.RLock()
	s.gauges.lock.RLock()
	defer s.counters.lock.RUnlock()
	defer s.gauges.lock.RUnlock()

	return len(s.counters.storage) + len(s.gauges.storage), nil
}

func (s *MemStorage) GetMetric(ctx context.Context, mType string, mName string) (models.Metric, error) {
	var err = storage.ErrNoMetric
	metric := models.Metric{Type: mType, Name: mName}

	switch mType {
	case models.CounterTypeName:
		s.counters.lock.RLock()
		defer s.counters.lock.RUnlock()
		var ok bool
		metric.Delta, ok = s.counters.storage[mName]
		if ok {
			return metric, nil
		}
	case models.GaugeTypeName:
		s.gauges.lock.RLock()
		defer s.gauges.lock.RUnlock()
		var ok bool
		metric.Value, ok = s.gauges.storage[mName]
		if ok {
			return metric, nil
		}
	}

	return metric, err
}

func (s *MemStorage) UpdateMetric(ctx context.Context, in *models.Metric) (models.Metric, error) {
	metric := models.Metric{Type: in.Type, Name: in.Name}

	switch metric.Type {
	case models.CounterTypeName:
		s.counters.lock.Lock()
		s.counters.storage[in.Name] += in.Delta
		counter := s.counters.storage[in.Name]
		s.counters.lock.Unlock()

		metric.Delta = counter
	case models.GaugeTypeName:
		s.gauges.lock.Lock()
		s.gauges.storage[in.Name] = in.Value
		gauge := s.gauges.storage[in.Name]
		s.gauges.lock.Unlock()

		metric.Value = gauge
	default:
		return metric, fmt.Errorf("unknown metric type: %s", metric.Type)
	}

	// is saveInterval = 0, than save metrics in synchronously
	if s.saveInterval == 0 {
		if err := s.save(); err != nil {
			logger.Slog.Errorw("metrics saving failed", "error", err.Error())
			return metric, err
		}
		logger.Slog.Debugw("metrics saved", "metric", in)
	}

	return metric, nil
}

func (s *MemStorage) ListMetric(ctx context.Context) ([]models.Metric, error) {
	s.counters.lock.RLock()
	defer s.counters.lock.RUnlock()
	s.gauges.lock.RLock()
	defer s.gauges.lock.RUnlock()

	metrics := make([]models.Metric, 0, len(s.counters.storage)+len(s.gauges.storage))

	for name, counter := range s.counters.storage {
		metrics = append(metrics, models.Metric{Name: name, Type: models.CounterTypeName, Delta: counter})
	}

	for name, gauge := range s.gauges.storage {
		metrics = append(metrics, models.Metric{Name: name, Type: models.GaugeTypeName, Value: gauge})
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Name < metrics[j].Name
	})

	return metrics, nil
}
