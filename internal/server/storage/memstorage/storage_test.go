package memstorage

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/server/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Create storage that save data in-memory only
// At the end of running test close the server
func newInMemory(t *testing.T) *MemStorage  {
	t.Helper()

	storage, err := New("", 0, false)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := storage.Close()
		require.NoError(t, err)
	})

	return storage
}

// Create new storage with persistent synchronous file storage.
// Return filename of persistent file and server itself
// At the end of test close server and delete temp file
func newPersistentSync(t *testing.T) (*MemStorage, string) {
	t.Helper()

	// Tmp directory for persistent storage
	tmpFile, err := os.CreateTemp("", "sync_metrics_*.json")
	require.NoError(t, err)
	filename := tmpFile.Name()
	t.Cleanup(func() { _ = os.Remove(filename) })

	storage, err := New(filename, 0, false)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := storage.Close()
		require.NoError(t, err)
	})

	return storage, filename
}

func TestMemStorage_UpdateMetric(t *testing.T) {
	mCounter := models.Metric{Name: "foo", Type: models.CounterTypeName, Delta: 10}
	mGauge := models.Metric{Name: "foo", Type: models.GaugeTypeName, Value: 500.23}

	t.Run("counter update once ok", func(t *testing.T) {
		s := newInMemory(t)

		got, err := s.UpdateMetric(context.TODO(), &mCounter)

		assert.NoError(t, err)
		assert.Equal(t, mCounter, got)
	})

	t.Run("counter update several ok", func(t *testing.T) {
		s := newInMemory(t)

		_, _ = s.UpdateMetric(context.TODO(), &mCounter)
		got, err := s.UpdateMetric(context.TODO(), &mCounter)

		assert.NoError(t, err)
		assert.Equal(t, models.Metric{Type: models.CounterTypeName, Name: "foo", Delta: 20}, got, "counter should increase")
	})

	t.Run("gauge update once ok", func(t *testing.T) {
		s := newInMemory(t)

		got, err := s.UpdateMetric(context.TODO(), &mGauge)

		assert.NoError(t, err)
		assert.Equal(t, mGauge, got)
	})

	t.Run("gauge update several ok", func(t *testing.T) {
		s := newInMemory(t)
		yaGauge := models.Metric{Type: models.GaugeTypeName, Name: "foo", Value: 123.1}

		_, _ = s.UpdateMetric(context.TODO(), &mGauge)
		got, err := s.UpdateMetric(context.TODO(), &yaGauge)

		assert.NoError(t, err)
		assert.Equal(t, yaGauge, got, "Gauge on update should replace")
	})

	t.Run("fail if unknown type", func(t *testing.T) {
		s := newInMemory(t)
		metric := models.Metric{Type: "unknown", Name: "foo", Value: 500.23}

		_, err := s.UpdateMetric(context.TODO(), &metric)

		require.Error(t, err)
	})

	t.Run("concurrently ok", func(t *testing.T) {
		s := newInMemory(t)

		var wg sync.WaitGroup
		for range 10 {
			wg.Add(1)
			go func() {
				_, _ = s.UpdateMetric(context.TODO(), &mCounter)
				_, _ = s.UpdateMetric(context.TODO(), &mGauge)
				wg.Done()
			}()
		}
		got, err := s.UpdateMetric(context.TODO(), &mGauge)

		require.NoError(t, err)
		assert.Equal(t, mGauge, got)
	})

	t.Run("sync save file ok", func(t *testing.T) {
		s, filename := newPersistentSync(t)

		_, err := s.UpdateMetric(t.Context(), &mCounter)
		require.NoError(t, err)

		data, err := os.ReadFile(filename)
		require.NoError(t, err)
		require.JSONEq(t, `
			[{
				"id": "foo",
				"type": "counter",
				"delta": 10
			}]`,
			string(data),
		)
	})
}

func TestMemStorage_CountMetric(t *testing.T) {
	s := newInMemory(t)
	_, _ = s.UpdateMetric(context.TODO(), &models.Metric{Name: "foo", Type: models.CounterTypeName, Delta: 10})
	_, _ = s.UpdateMetric(context.TODO(), &models.Metric{Name: "bar", Type: models.CounterTypeName, Delta: 200})
	_, _ = s.UpdateMetric(context.TODO(), &models.Metric{Name: "goo", Type: models.GaugeTypeName, Value: 500.233})

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() { s.CountMetric(context.TODO()); wg.Done() }() // nolint:errcheck
	}
	wg.Wait()
	got, err := s.CountMetric(context.TODO())

	assert.NoError(t, err)
	assert.Equal(t, 3, got)
}

func TestMemStorage_GetMetric(t *testing.T) {
	fooCounter := models.Metric{Name: "foo", Type: models.CounterTypeName, Delta: 10}
	barCounter := models.Metric{Name: "bar", Type: models.CounterTypeName, Delta: 200}
	fooGauge := models.Metric{Name: "foo", Type: models.GaugeTypeName, Value: 500.233}

	s := newInMemory(t)
	_, _ = s.UpdateMetric(context.TODO(), &fooCounter)
	_, _ = s.UpdateMetric(context.TODO(), &barCounter)
	_, _ = s.UpdateMetric(context.TODO(), &fooGauge)

	t.Run("sync ok", func(t *testing.T) {
		type expected struct {
			metric models.Metric
			err    error
		}
		type args struct {
			mType string
			mName string
		}
		tCases := []struct {
			name     string
			args     args
			expected expected
		}{
			{
				"get counter ok",
				args{models.CounterTypeName, "foo"},
				expected{fooCounter, nil},
			},
			{
				"get gauge ok",
				args{models.GaugeTypeName, "foo"},
				expected{fooGauge, nil},
			},
			{
				"get counter bad",
				args{models.CounterTypeName, "unknown"},
				expected{models.Metric{Type: models.CounterTypeName, Name: "unknown"}, storage.ErrNoMetric},
			},
			{
				"get gauge bad",
				args{models.GaugeTypeName, "bar"}, // existed id but for counter only
				expected{models.Metric{Type: models.GaugeTypeName, Name: "bar"}, storage.ErrNoMetric},
			},
		}

		for _, tc := range tCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := s.GetMetric(context.TODO(), tc.args.mType, tc.args.mName)

				require.Equal(t, tc.expected.err, err)
				assert.EqualValues(t, tc.expected.metric, got)
			})
		}

	})

	t.Run("concurrently ok", func(t *testing.T) {
		var wg sync.WaitGroup
		for range 10 {
			wg.Add(1)
			go func() {
				s.GetMetric(context.TODO(), models.CounterTypeName, "foo") // nolint:errcheck
				s.GetMetric(context.TODO(), models.GaugeTypeName, "foo")   // nolint:errcheck
				wg.Done()
			}()
		}
		wg.Wait()
		got, err := s.GetMetric(context.TODO(), models.GaugeTypeName, "foo")

		assert.NoError(t, err)
		assert.EqualValues(t, fooGauge, got)
	})
}

func TestMemStorage_ListMetric(t *testing.T) {
	fooCounter := models.Metric{Name: "foo", Type: models.CounterTypeName, Delta: 10}
	barCounter := models.Metric{Name: "bar", Type: models.CounterTypeName, Delta: 200}
	fooGauge := models.Metric{Name: "foo", Type: models.GaugeTypeName, Value: 500.233}

	s := newInMemory(t)
	_, _ = s.UpdateMetric(context.TODO(), &fooCounter)
	_, _ = s.UpdateMetric(context.TODO(), &barCounter)
	_, _ = s.UpdateMetric(context.TODO(), &fooGauge)

	t.Run("list ok", func(t *testing.T) {
		metrics, err := s.ListMetric(context.TODO())
		require.NoError(t, err)

		require.Equal(t, 3, len(metrics))
		assert.ElementsMatch(t, []models.Metric{fooCounter, barCounter, fooGauge}, metrics)
	})

	t.Run("list concurrent ok", func(t *testing.T) {
		mResults := make([][]models.Metric, 3)

		// Run something to imitate concurrent access
		var wg sync.WaitGroup
		for idx := range mResults {
			wg.Add(1)
			go func() {
				mResults[idx], _ = s.ListMetric(context.TODO())
				wg.Done()
			}()
		}

		wg.Wait()
		assert.Equal(t, 3, len(mResults[0]))
		assert.Equal(t, 3, len(mResults[1]))
		assert.Equal(t, 3, len(mResults[2]))
	})
}

func TestMemStorage_UpdateMetricBulk(t *testing.T) {
	metrics := []models.Metric{
		{Name: "bar", Type: models.GaugeTypeName, Value: 431.10},
		{Name: "foo", Type: models.CounterTypeName, Delta: 10},
	}

	t.Run("update metric counter and gauge bulk ok", func(t *testing.T) {
		s := newInMemory(t)

		got, err := s.UpdateMetricBulk(context.TODO(), metrics)

		assert.NoError(t, err)
		assert.Equal(t, metrics, got)

		inMemory, err := s.ListMetric(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, inMemory, got)
	})

	t.Run("fail if unknown type", func(t *testing.T) {
		s := newInMemory(t)
		invalid := append(metrics, models.Metric{Name: "unknown", Type: "unknown"})

		got, err := s.UpdateMetricBulk(context.TODO(), invalid)

		assert.Error(t, err)
		assert.Equal(t, invalid, got)

		inMemory, err := s.ListMetric(context.TODO())
		require.NoError(t, err)
		assert.Equal(t, make([]models.Metric, 0), inMemory)
	})

	t.Run("save file sync ok", func(t *testing.T) {
		s, filename := newPersistentSync(t)

		// Create server with sync save, update any metric and stop
		_, err := s.UpdateMetricBulk(t.Context(), metrics)
		require.NoError(t, err)

		// Check file saved and data as expected
		data, err := os.ReadFile(filename)
		require.NoError(t, err)
		require.JSONEq(t, `
			[
				{
					"id": "bar",
					"type": "gauge",
					"value": 431.10
				},
				{
					"id": "foo",
					"type": "counter",
					"delta": 10
				}
			]`,
			string(data),
		)
	})
}
