package memstorage

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
)

// Save func for MemStorage
// Should write all the metrics to a storage's file as valid json
func memSave(s *MemStorage) error {
	s.fileLock.Lock()
	defer s.fileLock.Unlock()

	var err error
	if _, err = s.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if err = s.file.Truncate(0); err != nil {
		return err
	}

	buf := bufio.NewWriter(s.file)
	metrics, _ := s.ListMetric(context.TODO())

	err = json.NewEncoder(buf).Encode(metrics)
	if err != nil {
		return err
	}

	if err = buf.Flush(); err != nil {
		return err
	}

	return err
}
