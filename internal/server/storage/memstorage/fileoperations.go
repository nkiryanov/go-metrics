package memstorage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/nkiryanov/go-metrics/internal/models"
)

func (s *MemStorage) loadFromFile() (err error) {
	file, err := os.Open(s.filename)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist): // if file not exists, then nothing to restore
			return nil
		default:
			return err
		}
	}
	defer func() {
		err = file.Close()
	}()

	var metrics []models.Metric
	if err := json.NewDecoder(file).Decode(&metrics); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, m := range metrics {
		switch m.Type {
		case models.CounterTypeName:
			s.counters[m.Name] = m.Delta
		case models.GaugeTypeName:
			s.gauges[m.Name] = m.Value
		}
	}

	return nil
}

// Save in memory data to file
func (s *MemStorage) saveToFile() (err error) {
	// Do nothing if filename empty
	if s.filename == "" {
		return nil
	}

	metrics, err := s.ListMetric(context.TODO())
	if err != nil {
		return err
	}

	// Open file for write, overwrite if have data already
	file, err := os.OpenFile(s.filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
	}()

	return json.NewEncoder(file).Encode(metrics)
}

// Run job to save in memory data to file until stopped via <-s.stopCh
func (s *MemStorage) backgroundSaver() {
	ticker := time.NewTicker(s.saveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = s.saveToFile() // maybe goog idea to log
		case <-s.stopCh:
			return
		}
	}
}
