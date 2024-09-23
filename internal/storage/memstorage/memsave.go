package memstorage

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/nkiryanov/go-metrics/internal/models"
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
	encoder := json.NewEncoder(buf)

	// Write the opening of the JSON array manually
	if _, err = buf.WriteString("[\n"); err != nil {
		return err
	}

	// Iterate over metrics and write them as JSON
	len := s.Count()
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
