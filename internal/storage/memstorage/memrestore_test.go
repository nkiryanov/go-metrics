package memstorage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_memRestore(t *testing.T) {
	file, err := os.CreateTemp("", "saved_metrics_*.json")
	require.NoError(t, err)
	filename := file.Name()
	defer os.Remove(filename)

	// Write some metrics to file
	metricsJSON := `[
		{
			"id": "foo",
			"type": "counter",
			"delta": 10,
			"value": 0.0
		}
	]`
	err = os.WriteFile(filename, []byte(metricsJSON), 0644)
	require.NoError(t, err)

	t.Run("restore func ok", func(t *testing.T) {
		s, err := New(filename, 3*time.Minute, false) // Restore = false
		require.NoError(t, err)
		defer s.Close()

		err = memRestore(s)

		require.NoError(t, err)
		count, err := s.CountMetric(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}
