package memstatscollector

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nkiryanov/go-metrics/internal/models"
)

var (
	expectedStatNames = []string{
		// Gauges
		Alloc,
		BuckHashSys,
		Frees,
		GCCPUFraction,
		GCSys,
		HeapAlloc,
		HeapIdle,
		HeapInuse,
		HeapObjects,
		HeapReleased,
		HeapSys,
		LastGC,
		Lookups,
		MCacheInuse,
		MCacheSys,
		MSpanInuse,
		MSpanSys,
		Mallocs,
		NextGC,
		NumForcedGC,
		NumGC,
		OtherSys,
		PauseTotalNs,
		StackInuse,
		StackSys,
		Sys,
		TotalAlloc,
		RandomValue,
		// Counters
		PollCount,
	}
)

func TestMemStatsCollector(t *testing.T) {
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

		require.Len(t, got, len(expectedStatNames))
		require.EqualValues(t, expectedStatNames, extractNames(got))
	})

	// Do not run tests with -race flag
	t.Run("collect and save not race", func(t *testing.T) {
		var wg sync.WaitGroup
		c := New()

		for range 5 {
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

		assert.Len(t, got, len(expectedStatNames))
	})
}

func Benchmark_StatsCollector(b *testing.B) {
	collector := New()

	for b.Loop() {
		_ = collector.Collect(b.Context())
	}

}
