package memstorage

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/nkiryanov/go-metrics/internal/models"
)

func (s *MemStorage) save() error {
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

func (s *MemStorage) restore() error {
	s.fileLock.Lock()
	defer s.fileLock.Unlock()
	var err error

	// Be sure to read from the beginning of the file
	if _, err = s.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	buf := bufio.NewReader(s.file)
	decoder := json.NewDecoder(buf)

	// Read first token. Expecting [ as list of metrics
	if _, err = decoder.Token(); err != nil {
		// File may be empty. In that case no metrics to restore and no error
		if err == io.EOF {
			return nil
		}
		return err
	}

	// Read and load metrics in storage
	metric := models.Metric{}
	for decoder.More() {
		if err = decoder.Decode(&metric); err != nil {
			return err
		}

		if _, err = s.UpdateMetric(&metric); err != nil {
			return err
		}
	}
	return nil
}
