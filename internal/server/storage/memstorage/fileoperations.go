package memstorage

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/nkiryanov/go-metrics/internal/models"
)

func (s *MemStorage) loadFromFile() (err error) {
	if s.file == nil {
		return nil // do nothing if persistent storage not set
	}

	s.fileLock.Lock()
	defer s.fileLock.Unlock()

	var metrics []models.Metric
	err = json.NewDecoder(s.file).Decode(&metrics)
	if err != nil {
		switch {
		case errors.Is(err, io.EOF): // empty file is ok, just restore nothing
			return nil
		default:
			return err
		}
	}

	s.update(metrics)

	return nil
}

// Save in memory data to file
func (s *MemStorage) saveToFile() error {
	// Do nothing if filename empty
	if s.file == nil {
		return nil
	}

	s.fileLock.Lock()
	defer s.fileLock.Unlock()

	// Truncate file to write new data
	_, err := s.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	err = s.file.Truncate(0)
	if err != nil {
		return err
	}

	metrics, err := s.ListMetric(context.TODO())
	if err != nil {
		return err
	}
	buf := bufio.NewWriter(s.file)

	err = json.NewEncoder(buf).Encode(metrics)
	if err != nil {
		return err
	}

	err = buf.Flush()
	if err != nil {
		return err
	}

	return nil
}

// Run job to save in memory data to file until stopped via <-s.stopCh
func (s *MemStorage) backgroundSaver() {
	ticker := time.NewTicker(s.saveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := s.saveToFile() // maybe goog idea to log
			if err != nil {
				s.lgr.Error("Metrics save failed", "err", err)
			} else {
				s.lgr.Info("Metrics saved to file")
			}
		case <-s.stopCh:
			return
		}
	}
}
