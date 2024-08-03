package storage

import (
	"sync"
)

type Storage interface {
	UpdateCounter(metric MetricName, value Countable) Countable
	GetCounter(metric MetricName) (Countable, bool)
	SetGauge(metric MetricName, value Gaugeable) Gaugeable
	GetGauge(metric MetricName) (Gaugeable, bool)
}

type (
	MetricType string
	MetricName string
	Gaugeable  float64
	Countable  int64
)

const (
	GaugeTypeName   MetricType = "gauge"
	CounterTypeName MetricType = "counter"
)

type counterStore struct {
	lock    sync.RWMutex
	storage map[MetricName]Countable
}

func (s *counterStore) UpdateMetric(metric MetricName, value Countable) Countable {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.storage[metric] += value
	return s.storage[metric]
}

func (s *counterStore) GetMetric(metric MetricName) (Countable, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	value, ok := s.storage[metric]
	return value, ok
}

type gaugeStore struct {
	lock    sync.RWMutex
	storage map[MetricName]Gaugeable
}

func (s *gaugeStore) UpdateMetric(metric MetricName, value Gaugeable) Gaugeable {
	s.lock.Lock()
	s.storage[metric] = value
	s.lock.Unlock()

	return value
}

func (s *gaugeStore) GetMetric(metric MetricName) (Gaugeable, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	value, ok := s.storage[metric]
	return value, ok
}

type MemStorage struct {
	gauges   gaugeStore
	counters counterStore
}

func (s *MemStorage) UpdateCounter(metric MetricName, value Countable) Countable {
	return s.counters.UpdateMetric(metric, value)
}

func (s *MemStorage) GetCounter(metric MetricName) (Countable, bool) {
	return s.counters.GetMetric(metric)
}

func (s *MemStorage) SetGauge(metric MetricName, value Gaugeable) Gaugeable {
	return s.gauges.UpdateMetric(metric, value)
}

func (s *MemStorage) GetGauge(metric MetricName) (Gaugeable, bool) {
	return s.gauges.GetMetric(metric)
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: counterStore{storage: make(map[MetricName]Countable)},
		gauges:   gaugeStore{storage: make(map[MetricName]Gaugeable)},
	}
}
