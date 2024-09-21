package memstorage

import (
	"os"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper. Create temp directory and defer fn that delete it.
func tmpFilename(t *testing.T) (filename string, deferFn func()) {
	// Tmp directory for persistent storage
	file, err := os.CreateTemp("", "save_restore_tests")
	require.NoError(t, err)

	deferFn = func() {
		_ = os.Remove(file.Name())
	}

	return file.Name(), deferFn
}

func TestMemStorage_save(t *testing.T) {
	filename, deferFn := tmpFilename(t)
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
		data, err := os.ReadFile(mems.file.Name())
		require.NoError(t, err)
		content := string(data)
		assert.JSONEq(t, expectedJSON, content)
	})
}

func TestMemStorage_restore(t *testing.T) {
	filename, deferFn := tmpFilename(t)
	defer deferFn()

	// Write some metrics to file
	metricsJSON := `[
		{
			"id": "foo",
			"type": "counter",
			"delta": 10,
			"value": 0.0
		}
	]`
	err := os.WriteFile(filename, []byte(metricsJSON), 0644)
	require.NoError(t, err)

	t.Run("restore func ok", func(t *testing.T) {
		mems, err := New(filename, 3*time.Minute, false) // Restore = false
		require.NoError(t, err)
		defer mems.Close()

		err = mems.restore()

		require.NoError(t, err)
		assert.Equal(t, 1, mems.Count())
	})
}
