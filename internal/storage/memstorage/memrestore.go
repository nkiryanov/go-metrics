package memstorage

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/nkiryanov/go-metrics/internal/models"
)

// Restore func for MemStorage
// Should read storage file as list of metrics and write them to storage
func memRestore(s *MemStorage) error {
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
