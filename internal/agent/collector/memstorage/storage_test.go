package memstorage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nkiryanov/go-metrics/internal/models"
)

func TestMemStorage(t *testing.T) {
	metrics := []models.Metric{
		{Name: "cpu", Type: "gauge", Value: 23.22},
		{Name: "PollCount", Type: "counter", Delta: 10},
	}

	t.Run("empty not fail", func(t *testing.T) {
		s := New()

		got := s.List()

		assert.Len(t, got, 0, "Empty storage has return zero length slice")
	})

	t.Run("set and list ok", func(t *testing.T) {
		s := New()

		s.Set(metrics...)
		got := s.List()

		require.Len(t, got, 2)
		require.Equal(t, metrics, got)
	})

	t.Run("set override existing", func(t *testing.T) {
		s := New()
		s.Set(metrics...)
		newPoll := models.Metric{Name: "PollCount", Type: "counter", Delta: 5}

		s.Set(newPoll)
		got := s.List()

		require.Len(t, got, 1, "Set has to entirely override existing metrics with new ones")
		require.Equal(t, []models.Metric{newPoll}, got)
	})

	// Do not forget to run tests with -race flag
	t.Run("not race", func(t *testing.T) {
		s := New()
		wg := sync.WaitGroup{}

		for range 10 {
			wg.Go(func() {
				s.Set(metrics...)
				_ = s.List()
			})
		}

		wg.Wait()
		got := s.List()

		assert.Equal(t, metrics, got)
	})
}
