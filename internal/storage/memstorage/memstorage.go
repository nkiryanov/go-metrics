package memstorage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

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
	gauges   gaugeStore
	counters counterStore

	file         *os.File
	fileLock     sync.Mutex
	saveInterval time.Duration
}

func New(filePath string, interval time.Duration, restore bool) (*MemStorage, error) {
	var err error
	var file *os.File

	// Open file as persistent storage for metrics
	file, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	storage := &MemStorage{
		counters: counterStore{storage: make(map[string]int64)},
		gauges:   gaugeStore{storage: make(map[string]float64)},

		file:         file,
		saveInterval: interval,
	}

	if !restore {
		return storage, nil
	}

	// Restore storage data from file
	if err = storage.restore(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *MemStorage) Close() error {
	if err := s.save(); err != nil {
		return err
	}
	return s.file.Close()
}

func (s *MemStorage) restore() error {
	s.fileLock.Lock()
	defer s.fileLock.Unlock()
	var err error

	// Be sure to read from the beginning of the file
	if _, err = s.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	buf := bufio.NewReader(s.file)
	decoder := json.NewDecoder(buf)

	// Read first token. Expecting [ as list of metrics
	if _, err = decoder.Token(); err != nil {
		// File may be empty. In that case no metrics to restore and no error
		if err == io.EOF {
			return nil
		}
		return err
	}

	// Read and load metrics in storage
	metric := models.Metric{}
	for decoder.More() {
		if err = decoder.Decode(&metric); err != nil {
			return err
		}

		if _, err = s.UpdateMetric(&metric); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemStorage) save() error {
	s.fileLock.Lock()
	defer s.fileLock.Unlock()

	var err error
	if _, err = s.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err = s.file.Truncate(0); err != nil {
		return err
	}

	// Iterate over metrics and write to file
	len := s.Count()
	buf := bufio.NewWriter(s.file)
	encoder := json.NewEncoder(buf)

	// Write the opening of the JSON array manually
	if _, err = buf.WriteString("[\n"); err != nil {
		return err
	}

	// Iterate over metrics and write them as JSON
	if err = s.Iterate(func(m models.Metric) error {
		var err error

		if err = encoder.Encode(m); err != nil {
			return err
		}
		len--
		if len > 0 {
			if _, err = buf.WriteString(","); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// Write the closing of the JSON array manually
	if _, err = buf.WriteString("]"); err != nil {
		return err
	}

	if err = buf.Flush(); err != nil {
		return err
	}

	return nil
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
