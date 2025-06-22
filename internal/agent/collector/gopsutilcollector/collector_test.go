package gopsutilcollector

import (
	"sync"
	"testing"

	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGopsutilCollector(t *testing.T) {
	extractNames := func(metrics []models.Metric) []string {
		names := make([]string, 0, len(metrics))
		for _, m := range metrics {
			names = append(names, m.Name)
		}
		return names
	}

	t.Run("stats initially empty", func(t *testing.T) {
		c := New()

		got := c.List()

		require.Equal(t, []models.Metric{}, got)
	})

	t.Run("collect stats ok", func(t *testing.T) {
		c := New()

		err := c.Collect(t.Context())
		require.NoError(t, err)
		got := c.List()

		require.GreaterOrEqual(t, len(got), 3, "Expect TotalMemory, FreeMemory and at least one CpuUtilization metric")
		assert.Contains(t, extractNames(got), "TotalMemory")
		assert.Contains(t, extractNames(got), "FreeMemory")
		assert.Contains(t, extractNames(got), "CPUutilization1")
	})

	// Do not run tests with -race flag
	t.Run("collect and save not race", func(t *testing.T) {
		var wg sync.WaitGroup
		c := New()

		for range 3 {
			wg.Add(2)
			go func() {
				defer wg.Done()
				err := c.Collect(t.Context())
				require.NoError(t, err)
			}()
			go func() {
				defer wg.Done()
				c.List()
			}()
		}

		wg.Wait()
		got := c.List()

		require.GreaterOrEqual(t, len(got), 3, "Expect TotalMemory, FreeMemory and at least one CpuUtilization metric")
	})
}
