package memstorage

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
)

func getFile(t *testing.T) string {
	t.Helper()

	file, err := os.CreateTemp("", "metrics_*.json")
	require.NoError(t, err)
	filename := file.Name()

	t.Cleanup(func() {
		err = os.Remove(filename)
		require.True(t, err == nil || errors.Is(err, os.ErrNotExist), "File deletion ok or it removed already")
	})

	return filename
}

func TestPgStorage_RestoreOption(t *testing.T) {

	t.Run("restore not existed file ok", func(t *testing.T) {
		notExisted := getFile(t)
		err := os.Remove(notExisted) // Remove file to make it not existed :)
		require.NoError(t, err)

		storage, err := New(
			notExisted, // file removed already
			1*time.Minute,
			true, // restore True
			logger.NewNoOpLogger(),
		)

		require.NoError(t, err, "Restore should ok if file not exists")

		err = storage.Close()
		require.NoError(t, err)
		assert.FileExists(t, notExisted, "File must be created on exit")
	})

	t.Run("restore empty file ok", func(t *testing.T) {
		emptyFile := getFile(t)

		storage, err := New(
			emptyFile,
			1*time.Minute,
			true, // restore True
			logger.NewNoOpLogger(),
		)

		require.NoError(t, err, "Restore empty must be ok")
		count, err := storage.CountMetric(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, 0, count)

		err = storage.Close()
		require.NoError(t, err)
		data, err := os.ReadFile(emptyFile)
		require.NoError(t, err)
		require.JSONEq(t, `[]`, string(data))

	})

	t.Run("restore file with data ok", func(t *testing.T) {
		filename := getFile(t)
		err := os.WriteFile(
			filename,
			[]byte(`[
				{"id": "foo", "type": "counter", "delta": 24},
				{"id": "foo", "type": "gauge", "value": 5.11},
				{"id": "bar", "type": "gauge", "value": 4.23}
			]`),
			0644,
		)
		require.NoError(t, err)

		storage, err := New(
			filename,
			1*time.Minute,
			true, // restore True
			logger.NewNoOpLogger(),
		)

		require.NoError(t, err, "Restore should ok")
		count, err := storage.CountMetric(t.Context())
		require.NoError(t, err)
		assert.Equal(t, 3, count)

		err = storage.Close()
		require.NoError(t, err)
	})

	t.Run("save by interval ok", func(t *testing.T) {
		filename := getFile(t)

		// Keep interval long enough to allow filesystem sync, otherwise reading may race and return unexpected values
		storage, err := New(
			filename,
			300*time.Millisecond,
			false,
			logger.NewNoOpLogger(),
		)
		require.NoError(t, err, "Storage start with save interval should start ok")

		_, err = storage.UpdateMetric(t.Context(), models.Metric{Name: "amount", Type: "counter", Delta: 3})
		require.NoError(t, err)

		// Check file content before save interval happen
		data, err := os.ReadFile(filename)
		require.NoError(t, err)
		require.Equal(t, "", string(data))

		// Check file content when time interval pass
		time.Sleep(400 * time.Millisecond)
		data, err = os.ReadFile(filename)
		require.NoError(t, err)
		assert.JSONEq(t, `[{"id": "amount", "type": "counter", "delta": 3}]`, string(data))
	})
}
