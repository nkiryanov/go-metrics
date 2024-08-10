package storage

import (
	"fmt"
	"strconv"
	"sync"
)

const (
	CounterTypeName = "counter"
	GaugeTypeName   = "gauge"
)

type Counter int64

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func (c Counter) Type() string {
	return CounterTypeName
}

type Gauge float64

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

func (g Gauge) Type() string {
	return GaugeTypeName
}

type counterStore struct {
	lock    sync.RWMutex
	storage map[string]Counter
}

type gaugeStore struct {
	lock    sync.RWMutex
	storage map[string]Gauge
}

type MemStorage struct {
	gauges   gaugeStore
	counters counterStore
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: counterStore{storage: make(map[string]Counter)},
		gauges:   gaugeStore{storage: make(map[string]Gauge)},
	}
}

func (s *MemStorage) GetCounter(mName string) (Counter, bool) {
	s.counters.lock.RLock()
	defer s.counters.lock.RUnlock()

	value, ok := s.counters.storage[mName]
	return value, ok
}

func (s *MemStorage) UpdateCounter(mName string, value Counter) Counter {
	s.counters.lock.Lock()
	defer s.counters.lock.Unlock()

	s.counters.storage[mName] += value
	return s.counters.storage[mName]
}

func (s *MemStorage) IterateCounters(fn func(mName string, mValue Counter)) {
	s.counters.lock.RLock()
	defer s.counters.lock.RUnlock()

	for mName, mValue := range s.counters.storage {
		fn(mName, mValue)
	}
}

func (s *MemStorage) GetGauge(mName string) (Gauge, bool) {
	s.gauges.lock.RLock()
	defer s.gauges.lock.RUnlock()

	value, ok := s.gauges.storage[mName]
	return value, ok
}

func (s *MemStorage) UpdateGauge(mName string, value Gauge) Gauge {
	s.gauges.lock.Lock()
	s.gauges.storage[mName] = value
	s.gauges.lock.Unlock()

	return value
}

func (s *MemStorage) IterateGauges(fn func(mName string, mValue Gauge)) {
	s.gauges.lock.RLock()
	defer s.gauges.lock.RUnlock()

	for mName, mValue := range s.gauges.storage {
		fn(mName, mValue)
	}
}

func (s *MemStorage) Len() int {
	s.counters.lock.RLock()
	defer s.counters.lock.RUnlock()

	s.gauges.lock.RLock()
	defer s.gauges.lock.RUnlock()

	return len(s.counters.storage) + len(s.gauges.storage)
}

func (s *MemStorage) GetMetric(mType string, mName string) (Storable, bool, error) {
	var val Storable
	var ok bool

	switch mType {
	case CounterTypeName:
		val, ok = s.GetCounter(mName)
	case GaugeTypeName:
		val, ok = s.GetGauge(mName)
	default:
		return nil, false, fmt.Errorf("unknown metric type: %s", mType)
	}

	return val, ok, nil
}

func (s *MemStorage) UpdateMetric(mName string, mValue Storable) (Storable, error) {
	switch storable := mValue.(type) {
	case Counter:
		return s.UpdateCounter(mName, storable), nil
	case Gauge:
		return s.UpdateGauge(mName, storable), nil
	default:
		return nil, fmt.Errorf("unknown metric type: %s", storable)
	}
}

func (s *MemStorage) Iterate(fn IterFunc) {
	s.counters.lock.RLock()
	for mName, mValue := range s.counters.storage {
		fn(mValue.Type(), mName, mValue)
	}
	s.counters.lock.RUnlock()

	s.gauges.lock.RLock()
	for mName, mValue := range s.gauges.storage {
		fn(mValue.Type(), mName, mValue)
	}
	s.gauges.lock.RUnlock()
}
