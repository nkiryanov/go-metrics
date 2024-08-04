package storage

import (
	"sort"
	"strconv"
	"sync"
)

type Storage interface {
	UpdateCounter(metric MetricName, value Countable) Countable
	GetCounter(metric MetricName) (Countable, bool)
	SetGauge(metric MetricName, value Gaugeable) Gaugeable
	GetGauge(metric MetricName) (Gaugeable, bool)
	IterateGauges(func(metric MetricName, value Gaugeable))
	IterateCounters(func(metric MetricName, value Countable))
	ListMetrics() []MetricRepr
}

type (
	MetricType string
	MetricName string
	Gaugeable  float64
	Countable  int64
	MetricRepr struct {
		mName  string
		mType  string
		mValue string
	}
)

const (
	CounterTypeName MetricType = "counter"
	GaugeTypeName   MetricType = "gauge"
)

func (c Countable) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func (g Gaugeable) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

type counterStore struct {
	lock    sync.RWMutex
	storage map[MetricName]Countable
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

type MemStorage struct {
	gauges   gaugeStore
	counters counterStore
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: counterStore{storage: make(map[MetricName]Countable)},
		gauges:   gaugeStore{storage: make(map[MetricName]Gaugeable)},
	}
}

func (s *MemStorage) UpdateCounter(metric MetricName, value Countable) Countable {
	s.counters.lock.Lock()
	defer s.counters.lock.Unlock()

	s.counters.storage[metric] += value
	return s.counters.storage[metric]
}

func (s *MemStorage) GetCounter(metric MetricName) (Countable, bool) {
	s.counters.lock.RLock()
	defer s.counters.lock.RUnlock()

	value, ok := s.counters.storage[metric]
	return value, ok
}

func (s *MemStorage) SetGauge(metric MetricName, value Gaugeable) Gaugeable {
	s.gauges.lock.Lock()
	s.gauges.storage[metric] = value
	s.gauges.lock.Unlock()

	return value
}

func (s *MemStorage) GetGauge(metric MetricName) (Gaugeable, bool) {
	s.gauges.lock.RLock()
	defer s.gauges.lock.RUnlock()

	value, ok := s.gauges.storage[metric]
	return value, ok
}

func (s *MemStorage) IterateCounters(fn func(metric MetricName, value Countable)) {
	s.counters.lock.RLock()
	defer s.counters.lock.RUnlock()

	for metric, value := range s.counters.storage {
		fn(metric, value)
	}
}

func (s *MemStorage) Len() int {
	s.counters.lock.RLock()
	defer s.counters.lock.RUnlock()
	s.gauges.lock.RLock()
	defer s.gauges.lock.RUnlock()

	return len(s.counters.storage) + len(s.gauges.storage)
}

func (s *MemStorage) IterateGauges(fn func(metric MetricName, value Gaugeable)) {
	s.gauges.lock.RLock()
	defer s.gauges.lock.RUnlock()

	for metric, value := range s.gauges.storage {
		fn(metric, value)
	}
}

func (s *MemStorage) ListMetrics() []MetricRepr {
	metrics := make([]MetricRepr, 0, s.Len())

	s.IterateCounters(func(name MetricName, value Countable) {
		metrics = append(metrics, MetricRepr{mType: string(CounterTypeName), mName: string(name), mValue: value.String()})
	})

	s.IterateGauges(func(metric MetricName, value Gaugeable) {
		metrics = append(metrics, MetricRepr{mType: string(GaugeTypeName), mName: string(metric), mValue: value.String()})
	})

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].mName < metrics[j].mName
	})

	return metrics
}
