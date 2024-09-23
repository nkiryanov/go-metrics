package storage

import (
	"fmt"
	"sync"

	"github.com/nkiryanov/go-metrics/internal/models"
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
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: counterStore{storage: make(map[string]int64)},
		gauges:   gaugeStore{storage: make(map[string]float64)},
	}
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

func (s *MemStorage) Iterate(fn IterFunc) {
	s.counters.lock.RLock()
	for id, counter := range s.counters.storage {
		fn(models.Metric{
			ID:    id,
			MType: models.CounterTypeName,
			Delta: counter,
		})
	}
	s.counters.lock.RUnlock()

	s.gauges.lock.RLock()
	for id, gauge := range s.gauges.storage {
		fn(models.Metric{
			ID:    id,
			MType: models.GaugeTypeName,
			Value: gauge,
		})
	}
	s.gauges.lock.RUnlock()
}
