package memstorage

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper. Create temp directory and defer fn that delete it.
func tmpFilename() (filename string, deferFn func()) {
	// Tmp directory for persistent storage
	tmpDir := os.TempDir()
	filename = tmpDir + "metrics_storage.json"

	deferFn = func() {
		_ = os.RemoveAll(tmpDir)
	}

	return filename, deferFn
}

func TestMemStorage_save(t *testing.T) {
	filename, deferFn := tmpFilename()
	defer deferFn()

	// Create storage and add some metrics
	mems, _ := New(filename, 3*time.Minute, true)
	defer mems.Close()
	_, _ = mems.UpdateMetric(&models.Metric{ID: "foo", MType: models.CounterTypeName, Delta: 10})
	_, _ = mems.UpdateMetric(&models.Metric{ID: "goo", MType: models.GaugeTypeName, Value: 500.233})

	t.Run("save ok", func(t *testing.T) {
		err := mems.save()
		expectedJSON := `[
			{
				"id": "foo",
				"type": "counter",
				"delta":10,
				"value": 0.0
			},
			{
				"id": "goo",
				"type": "gauge",
				"delta": 0,
				"value": 500.233
			}
		]`

		require.NoError(t, err)
		_, _ = mems.file.Seek(0, io.SeekStart)
		data, err := io.ReadAll(mems.file)
		require.NoError(t, err)
		content := string(data)
		assert.JSONEq(t, expectedJSON, content)
	})
}

func TestMemStorage_restore(t *testing.T) {
	tmpDir := os.TempDir()
	filepath := tmpDir + "metrics_storage.json"

	// On close storage save state to file
	mems, _ := New(filepath, 3*time.Minute, true)
	_, _ = mems.UpdateMetric(&models.Metric{ID: "foo", MType: models.CounterTypeName, Delta: 10})
	_, _ = mems.UpdateMetric(&models.Metric{ID: "goo", MType: models.GaugeTypeName, Value: 500.233})
	mems.Close()

	t.Run("restore ok", func(t *testing.T) {
		mems, _ = New(tmpDir+"metrics_storage.json", 3*time.Minute, true)

		err := mems.restore()

		require.NoError(t, err)
		assert.Equal(t, 2, mems.Count())
	})
}
