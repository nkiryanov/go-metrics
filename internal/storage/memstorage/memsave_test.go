package memstorage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_memSave(t *testing.T) {
	// Create temp file to save metrics
	file, err := os.CreateTemp("", "saved_metrics_*.json")
	require.NoError(t, err)
	filename := file.Name()
	defer os.Remove(filename) // nolint:errcheck

	// Create storage and add some metrics
	s, _ := New(filename, 3*time.Minute, true)
	defer s.Close() // nolint:errcheck
	_, _ = s.UpdateMetric(context.TODO(), &models.Metric{Name: "foo", Type: models.CounterTypeName, Delta: 10})
	_, _ = s.UpdateMetric(context.TODO(), &models.Metric{Name: "goo", Type: models.GaugeTypeName, Value: 500.233})

	t.Run("save actually save metrics", func(t *testing.T) {
		expectedJSON := `[
			{
				"id": "foo",
				"type": "counter",
				"delta":10
			},
			{
				"id": "goo",
				"type": "gauge",
				"value": 500.233
			}
		]`

		err := memSave(s)

		require.NoError(t, err)
		data, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.JSONEq(t, expectedJSON, string(data))
	})
}
