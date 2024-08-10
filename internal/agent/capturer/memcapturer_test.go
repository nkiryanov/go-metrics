package capturer

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getStatNames(stats []Stat) []string {
	names := make([]string, 0, len(stats))
	for _, stat := range stats {
		names = append(names, stat.Name)
	}
	return names
}

func TestMemCapturer(t *testing.T) {
	expectedNames := append(gauges, counters...)

	t.Run("capture return all stats", func(t *testing.T) {
		mc := NewMemCapturer()

		stats := mc.Capture()

		assert.EqualValues(t, expectedNames, getStatNames(stats))
	})

	t.Run("CaptureAndSave actually save, ok", func(t *testing.T) {
		mc := NewMemCapturer()

		mc.CaptureAndSave()

		assert.Equal(t, len(expectedNames), len(mc.stor))
	})

	t.Run("Last on empty, ok", func(t *testing.T) {
		mc := NewMemCapturer()

		stats := mc.Last()

		assert.Equal(t, 0, len(stats), "it stats not saved yet should return empty slice")
	})

	t.Run("Last when captured", func(t *testing.T) {
		mc := NewMemCapturer()
		mc.CaptureAndSave()

		stats := mc.Last()

		require.Equal(t, len(expectedNames), len(stats))
		assert.EqualValues(t, expectedNames, getStatNames(stats))
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
		stats := mc.Last()

		assert.Equal(t, len(expectedNames), len(stats))
	})
}
