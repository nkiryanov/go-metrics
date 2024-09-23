package capturer

import (
	"sync"
	"testing"

	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getMetricIDs(metrics []models.Metric) []string {
	ids := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		ids = append(ids, metric.ID)
	}
	return ids
}

func TestMemCapturer(t *testing.T) {
	expectedIDs := append(gauges, counters...)

	t.Run("capture return all stats", func(t *testing.T) {
		mc := NewMemCapturer()

		stats := mc.Capture()

		assert.EqualValues(t, expectedIDs, getMetricIDs(stats))
	})

	t.Run("CaptureAndSave actually save, ok", func(t *testing.T) {
		mc := NewMemCapturer()

		mc.CaptureAndSave()

		assert.Equal(t, len(expectedIDs), len(mc.stor))
		assert.Contains(t, mc.stor, models.Metric{ID: "PollCount", MType: "counter", Delta: 1}, "captured PollCount has to be on first call")
	})

	t.Run("Last on empty, ok", func(t *testing.T) {
		mc := NewMemCapturer()

		metrics := mc.Last()

		assert.Equal(t, 0, len(metrics), "should return empty slice if metrics not saved yet")
	})

	t.Run("Last when captured", func(t *testing.T) {
		mc := NewMemCapturer()
		mc.CaptureAndSave()

		metrics := mc.Last()

		require.Equal(t, len(expectedIDs), len(metrics))
		assert.EqualValues(t, expectedIDs, getMetricIDs(metrics))
	})

	t.Run("save and read not race", func(t *testing.T) {
		var wg sync.WaitGroup
		mc := NewMemCapturer()

		// Run concurrently to possible race conditions
		for range 5 {
			wg.Add(2)
			go func() { mc.CaptureAndSave(); wg.Done() }()
			go func() { mc.Last(); wg.Done() }()
		}

		wg.Wait()
		metrics := mc.Last()

		assert.Equal(t, len(expectedIDs), len(metrics))
	})
}
