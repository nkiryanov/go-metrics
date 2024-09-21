package memstorage

import (
	"context"
	"fmt"
	"os"
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

type MemStorage struct {
	// Metric storages
	gauges   gaugeStore
	counters counterStore

	// File as a persistent storage
	// Save interval means how often metrics should be saved (0 — should saved synchronously, on each update)
	file         *os.File
	fileLock     sync.Mutex
	saveInterval time.Duration

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
		go intervalSaver(storage)
		logger.Slog.Infow("storage: interval saver started", "interval", storage.saveInterval.String())
	}

	return storage, nil
}

func (s *MemStorage) Close() error {
	s.cancel()
	if err := s.save(); err != nil {
		return err
	}
	return s.file.Close()
}

func (s *MemStorage) Count() int {
	s.counters.lock.RLock()
	defer s.counters.lock.RUnlock()

	s.gauges.lock.RLock()
	defer s.gauges.lock.RUnlock()

	return len(s.counters.storage) + len(s.gauges.storage)
}

func (s *MemStorage) GetMetric(mID string, mType string) (models.Metric, bool, error) {
	var ok bool

	metric := models.Metric{ID: mID, MType: mType}

	switch mType {
	case models.CounterTypeName:
		s.counters.lock.RLock()
		var counter int64
		counter, ok = s.counters.storage[mID]
		metric.Delta = counter

		s.counters.lock.RUnlock()
	case models.GaugeTypeName:
		s.gauges.lock.RLock()
		var gauge float64
		gauge, ok = s.gauges.storage[mID]
		metric.Value = gauge

		s.gauges.lock.RUnlock()
	default:
		return metric, false, fmt.Errorf("unknown metric type: %s", mType)
	}

	return metric, ok, nil
}

func (s *MemStorage) UpdateMetric(in *models.Metric) (models.Metric, error) {
	metric := models.Metric{ID: in.ID, MType: in.MType}

	switch metric.MType {
	case models.CounterTypeName:
		s.counters.lock.Lock()
		s.counters.storage[in.ID] += in.Delta
		counter := s.counters.storage[in.ID]
		s.counters.lock.Unlock()

		metric.Delta = counter
	case models.GaugeTypeName:
		s.gauges.lock.Lock()
		s.gauges.storage[in.ID] = in.Value
		gauge := s.gauges.storage[in.ID]
		s.gauges.lock.Unlock()

		metric.Value = gauge
	default:
		return metric, fmt.Errorf("unknown metric type: %s", metric.MType)
	}

	if s.isSaveSync() {
		go func() {
			if err := s.save(); err != nil {
				logger.Slog.Errorw("fail saving failed", "error", err.Error())
			} else {
				logger.Slog.Debugw("memstorage: metrics saved", "metric", "in")
			}
		}()
	}

	return metric, nil
}

func (s *MemStorage) Iterate(fn storage.IterFunc) error {
	// Lock counters and gauges to be sure len will not change during iteration
	s.counters.lock.RLock()
	defer s.counters.lock.RUnlock()
	s.gauges.lock.RLock()
	defer s.gauges.lock.RUnlock()

	var err error
	for id, counter := range s.counters.storage {
		if err = fn(models.Metric{
			ID:    id,
			MType: models.CounterTypeName,
			Delta: counter,
		}); err != nil {
			return err
		}
	}

	for id, gauge := range s.gauges.storage {
		if err = fn(models.Metric{
			ID:    id,
			MType: models.GaugeTypeName,
			Value: gauge,
		}); err != nil {
			return err
		}
	}
	return nil
}
