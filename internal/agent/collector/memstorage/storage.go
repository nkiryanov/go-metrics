package memstorage

import (
	"sync"

	"github.com/nkiryanov/go-metrics/internal/models"
)

// Simple in memory metrics storage for agent's collectors
type MemStorage struct {
	metrics []models.Metric
	lock sync.RWMutex
}

func New() *MemStorage {
	return &MemStorage{
		metrics: make([]models.Metric, 0),
	}
}

func (s *MemStorage) Set(metrics ...models.Metric) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.metrics = append(s.metrics[:0], metrics...)
}

func (s *MemStorage) List() []models.Metric {
	s.lock.RLock()
	defer s.lock.RUnlock()

	metrics := make([]models.Metric, len(s.metrics))
	copy(metrics, s.metrics)

	return metrics
}
