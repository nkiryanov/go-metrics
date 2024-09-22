package memstorage

import (
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helper. Create storage that store state in temp file.
// Return *MemStorage and close func. The close should be run on end of the test, to release resources
func memstorage(t *testing.T, interval time.Duration) (*MemStorage, func()) {
	var err error

	// Tmp directory for persistent storage
	tmpFile, err := os.CreateTemp("", "metrics_*.json")
	require.NoError(t, err)
	filename := tmpFile.Name()

	storage, err := New(filename, interval, true)
	require.NoError(t, err)

	closeFn := func() {
		err = storage.Close()
		require.NoError(t, err)
		_ = os.Remove(filename)
	}

	return storage, closeFn
}

func TestMemStorage_UpdateMetric(t *testing.T) {
	mCounter := models.Metric{ID: "foo", MType: models.CounterTypeName, Delta: 10}
	mGauge := models.Metric{ID: "foo", MType: models.GaugeTypeName, Value: 500.23}

	t.Run("counter update once ok", func(t *testing.T) {
		storage, close := memstorage(t, 3*time.Minute)
		defer close()

		got, err := storage.UpdateMetric(&mCounter)

		assert.NoError(t, err)
		assert.Equal(t, mCounter, got)
	})

	t.Run("counter update several ok", func(t *testing.T) {
		storage, close := memstorage(t, 3*time.Minute)
		defer close()
		metric := models.Metric{ID: "foo", MType: models.CounterTypeName, Delta: 10}

		_, _ = storage.UpdateMetric(&metric)
		got, err := storage.UpdateMetric(&metric)

		assert.NoError(t, err)
		assert.Equal(t, models.Metric{ID: "foo", MType: models.CounterTypeName, Delta: 20}, got, "counter should increase")
	})

	t.Run("gauge update once ok", func(t *testing.T) {
		storage, close := memstorage(t, 3*time.Minute)
		defer close()

		got, err := storage.UpdateMetric(&mGauge)

		assert.NoError(t, err)
		assert.Equal(t, mGauge, got)
	})

	t.Run("gauge update several ok", func(t *testing.T) {
		storage, close := memstorage(t, 3*time.Minute)
		defer close()
		yaGauge := models.Metric{ID: "foo", MType: models.GaugeTypeName, Value: 123.1}

		_, _ = storage.UpdateMetric(&mGauge)
		got, err := storage.UpdateMetric(&yaGauge)

		assert.NoError(t, err)
		assert.Equal(t, yaGauge, got, "Gauge on update should replace")
	})

	t.Run("fail if unknown type", func(t *testing.T) {
		storage, close := memstorage(t, 3*time.Minute)
		defer close()
		metric := models.Metric{ID: "foo", MType: "unknown", Value: 500.23}

		_, err := storage.UpdateMetric(&metric)

		require.Error(t, err)
	})

	t.Run("concurrently ok", func(t *testing.T) {
		storage, close := memstorage(t, 3*time.Minute)
		defer close()

		var wg sync.WaitGroup
		for range 10 {
			wg.Add(1)
			go func() {
				_, _ = storage.UpdateMetric(&mCounter)
				_, _ = storage.UpdateMetric(&mGauge)
				wg.Done()
			}()
		}
		got, err := storage.UpdateMetric(&mGauge)

		require.NoError(t, err)
		assert.Equal(t, mGauge, got)
	})

	t.Run("call saver ok", func(t *testing.T) {
		tests := []struct {
			name         string
			interval     time.Duration
			expectCalled bool
		}{
			{
				"call if interval zero",
				0 * time.Second,
				true,
			},
			{
				"not call if interval > zero",
				1 * time.Second,
				false,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				storage, close := memstorage(t, tc.interval)
				defer close()

				// Mock memstorage saver
				var saverCalled atomic.Bool
				storage.saver = func(s *MemStorage) error { saverCalled.Store(true); return nil }

				// Update metric and give enough time to run goroutine
				_, err := storage.UpdateMetric(&mCounter)
				require.NoError(t, err)
				time.Sleep(100 * time.Millisecond)

				assert.Equal(t, tc.expectCalled, saverCalled.Load())
			})
		}
	})
}

func TestMemStorage_Count(t *testing.T) {
	storage, deferFn := memstorage(t, 3*time.Minute)
	defer deferFn()
	_, _ = storage.UpdateMetric(&models.Metric{ID: "foo", MType: models.CounterTypeName, Delta: 10})
	_, _ = storage.UpdateMetric(&models.Metric{ID: "bar", MType: models.CounterTypeName, Delta: 200})
	_, _ = storage.UpdateMetric(&models.Metric{ID: "goo", MType: models.GaugeTypeName, Value: 500.233})

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() { storage.Count(); wg.Done() }()
	}
	wg.Wait()
	got := storage.Count()

	assert.Equal(t, 3, got)
}

func TestMemStorage_GetMetric(t *testing.T) {
	fooCounter := models.Metric{ID: "foo", MType: models.CounterTypeName, Delta: 10}
	barCounter := models.Metric{ID: "bar", MType: models.CounterTypeName, Delta: 200}
	fooGauge := models.Metric{ID: "foo", MType: models.GaugeTypeName, Value: 500.233}

	storage, deferFn := memstorage(t, 3*time.Minute)
	defer deferFn()
	_, _ = storage.UpdateMetric(&fooCounter)
	_, _ = storage.UpdateMetric(&barCounter)
	_, _ = storage.UpdateMetric(&fooGauge)

	t.Run("get unexpected type", func(t *testing.T) {
		_, ok, err := storage.GetMetric("foo", "unknownType")

		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("get valid types", func(t *testing.T) {
		type expected struct {
			metric models.Metric
			ok     bool
		}
		type fnArgs struct {
			mID   string
			mType string
		}
		tCases := []struct {
			name     string
			fnArgs   fnArgs
			expected expected
		}{
			{
				"get counter ok",
				fnArgs{"foo", models.CounterTypeName},
				expected{fooCounter, true},
			},
			{
				"get gauge ok",
				fnArgs{"foo", models.GaugeTypeName},
				expected{fooGauge, true},
			},
			{
				"get counter bad",
				fnArgs{"unknown", models.CounterTypeName},
				expected{
					models.Metric{ID: "unknown", MType: models.CounterTypeName},
					false,
				},
			},
			{
				"get gauge bad",
				fnArgs{"bar", models.GaugeTypeName}, // existed id but for counter only
				expected{
					models.Metric{ID: "bar", MType: models.GaugeTypeName},
					false,
				},
			},
		}

		for _, tc := range tCases {
			t.Run(tc.name, func(t *testing.T) {
				got, ok, err := storage.GetMetric(tc.fnArgs.mID, tc.fnArgs.mType)

				require.NoError(t, err)
				assert.Equal(t, tc.expected.ok, ok)
				assert.EqualValues(t, tc.expected.metric, got)
			})
		}

	})

	t.Run("concurrently ok", func(t *testing.T) {
		var wg sync.WaitGroup
		for range 10 {
			wg.Add(1)
			go func() {
				storage.GetMetric("foo", models.CounterTypeName) // nolint:errcheck
				storage.GetMetric("foo", models.GaugeTypeName)   // nolint:errcheck
				wg.Done()
			}()
		}
		wg.Wait()
		got, ok, err := storage.GetMetric("foo", models.GaugeTypeName)

		assert.NoError(t, err)
		assert.True(t, ok)
		assert.EqualValues(t, fooGauge, got)
	})
}

func TestMemStorage_Iterate(t *testing.T) {
	fooCounter := models.Metric{ID: "foo", MType: models.CounterTypeName, Delta: 10}
	barCounter := models.Metric{ID: "bar", MType: models.CounterTypeName, Delta: 200}
	fooGauge := models.Metric{ID: "foo", MType: models.GaugeTypeName, Value: 500.233}

	storage, deferFn := memstorage(t, 3*time.Minute)
	defer deferFn()
	_, _ = storage.UpdateMetric(&fooCounter)
	_, _ = storage.UpdateMetric(&barCounter)
	_, _ = storage.UpdateMetric(&fooGauge)

	t.Run("iterate ok", func(t *testing.T) {
		metrics := make([]models.Metric, 0)

		_ = storage.Iterate(func(m models.Metric) error {
			metrics = append(metrics, m)
			return nil
		})

		require.Equal(t, 3, len(metrics))
		assert.ElementsMatch(t, []models.Metric{fooCounter, barCounter, fooGauge}, metrics)
	})

	t.Run("iterate concurrent ok", func(t *testing.T) {
		var iterCount = 0

		// Run something to imitate concurrent access
		var wg sync.WaitGroup
		defer wg.Wait()
		for range 10 {
			wg.Add(1)
			go func() {
				_ = storage.Iterate(func(models.Metric) error { return nil })
				wg.Done()
			}()
		}

		_ = storage.Iterate(func(models.Metric) error { iterCount++; return nil })
		assert.Equal(t, 3, iterCount)
	})

	t.Run("iterate stop on error", func(t *testing.T) {
		iterCalled := 0
		iterFn := func(m models.Metric) error {
			iterCalled++
			if iterCalled == 2 {
				return errors.New("error on iterated function occurred")
			}
			return nil
		}

		err := storage.Iterate(iterFn)

		require.Error(t, err)
		assert.Equal(t, 2, iterCalled, "Iterate should stop on error")
	})
}
