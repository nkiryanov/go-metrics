package storage

import (
	"fmt"
	"strconv"
	"sync"
)

// go:generate moq -out mocks/storage -pkg mocks -skip-ensure . Storage

type counterStore struct {
	lock    sync.RWMutex
	storage map[string]Countable
}

type gaugeStore struct {
	lock    sync.RWMutex
	storage map[string]Gaugeable
}

type MemStorage struct {
	gauges   gaugeStore
	counters counterStore
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: counterStore{storage: make(map[string]Countable)},
		gauges:   gaugeStore{storage: make(map[string]Gaugeable)},
	}
}

func (s *MemStorage) UpdateCounter(mName string, value Countable) Countable {
	s.counters.lock.Lock()
	defer s.counters.lock.Unlock()

	s.counters.storage[mName] += value
	return s.counters.storage[mName]
}

func (s *MemStorage) UpdateGauge(mName string, value Gaugeable) Gaugeable {
	s.gauges.lock.Lock()
	s.gauges.storage[mName] = value
	s.gauges.lock.Unlock()

	return value
}

func (s *MemStorage) UpdateValue(mType string, mName string, mValue string) (Storable, error) {
	switch mType {
	case CounterTypeName:
		countable, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return Countable(0), err
		}
		return s.UpdateCounter(mName, Countable(countable)), nil
	case GaugeTypeName:
		gauge, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return Gaugeable(0), err
		}
		return s.UpdateGauge(mName, Gaugeable(gauge)), nil
	default:
		return nil, fmt.Errorf("unknown metric type: %s", mType)
	}
}

func (s *MemStorage) GetCounter(mName string) (Countable, bool) {
	s.counters.lock.RLock()
	defer s.counters.lock.RUnlock()

	value, ok := s.counters.storage[mName]
	return value, ok
}

func (s *MemStorage) GetGauge(mName string) (Gaugeable, bool) {
	s.gauges.lock.RLock()
	defer s.gauges.lock.RUnlock()

	value, ok := s.gauges.storage[mName]
	return value, ok
}

func (s *MemStorage) IterateCounters(fn func(mName string, value Countable)) {
	s.counters.lock.RLock()
	defer s.counters.lock.RUnlock()

	for mName, mValue := range s.counters.storage {
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

func (s *MemStorage) IterateGauges(fn func(mName string, mValue Gaugeable)) {
	s.gauges.lock.RLock()
	defer s.gauges.lock.RUnlock()

	for mName, mValue := range s.gauges.storage {
		fn(mName, mValue)
	}
}

func (s *MemStorage) Iterate(fn func(mType string, mName string, mValue Storable)) {
	s.counters.lock.RLock()
	for mName, mValue := range s.counters.storage {
		fn(CounterTypeName, mName, mValue)
	}
	s.counters.lock.RUnlock()

	s.gauges.lock.RLock()
	for mName, mValue := range s.gauges.storage {
		fn(GaugeTypeName, mName, mValue)
	}
	s.gauges.lock.RUnlock()
}
