package capturer

import (
	"testing"
	"sync"

	"github.com/stretchr/testify/assert"
)

func getNames(stats []Stat) []string {
	names := make([]string, 0, len(stats))
	for _, stat := range stats {
		names = append(names, stat.Name)
	}
	return names
}

func TestMemCapturer_Capture(t *testing.T) {
	mc := NewMemCapturer()
	expectedNames := append(gauges, counters...)

	t.Run("capture return all stats", func(t *testing.T) {
		stats := mc.Capture()

		assert.EqualValues(t, expectedNames, getNames(stats))
	})

	t.Run("capture not race", func(t *testing.T) {
		var wg sync.WaitGroup

		// Run concurrently to possible race conditions
		for range 100 {
			wg.Add(1)
			go func() { mc.Capture(); wg.Done() }()
		}
		stats := mc.Capture()

		wg.Wait()
		assert.Equal(t, len(expectedNames), len(stats))
	})
}
