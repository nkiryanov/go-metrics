package memstorage

import (
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/storage"

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

	s, err := New(filename, interval, true)
	require.NoError(t, err)

	closeFn := func() {
		err = s.Close()
		require.NoError(t, err)
		_ = os.Remove(filename)
	}

	return s, closeFn
}

func TestMemStorage_UpdateMetric(t *testing.T) {
	mCounter := models.Metric{Name: "foo", Type: models.CounterTypeName, Delta: 10}
	mGauge := models.Metric{Name: "foo", Type: models.GaugeTypeName, Value: 500.23}

	t.Run("counter update once ok", func(t *testing.T) {
		s, close := memstorage(t, 3*time.Minute)
		defer close()

		got, err := s.UpdateMetric(&mCounter)

		assert.NoError(t, err)
		assert.Equal(t, mCounter, got)
	})

	t.Run("counter update several ok", func(t *testing.T) {
		s, close := memstorage(t, 3*time.Minute)
		defer close()
		metric := models.Metric{Type: models.CounterTypeName, Name: "foo", Delta: 10}

		_, _ = s.UpdateMetric(&metric)
		got, err := s.UpdateMetric(&metric)

		assert.NoError(t, err)
		assert.Equal(t, models.Metric{Type: models.CounterTypeName, Name: "foo", Delta: 20}, got, "counter should increase")
	})

	t.Run("gauge update once ok", func(t *testing.T) {
		s, close := memstorage(t, 3*time.Minute)
		defer close()

		got, err := s.UpdateMetric(&mGauge)

		assert.NoError(t, err)
		assert.Equal(t, mGauge, got)
	})

	t.Run("gauge update several ok", func(t *testing.T) {
		s, close := memstorage(t, 3*time.Minute)
		defer close()
		yaGauge := models.Metric{Type: models.GaugeTypeName, Name: "foo", Value: 123.1}

		_, _ = s.UpdateMetric(&mGauge)
		got, err := s.UpdateMetric(&yaGauge)

		assert.NoError(t, err)
		assert.Equal(t, yaGauge, got, "Gauge on update should replace")
	})

	t.Run("fail if unknown type", func(t *testing.T) {
		s, close := memstorage(t, 3*time.Minute)
		defer close()
		metric := models.Metric{Type: "unknown", Name: "foo", Value: 500.23}

		_, err := s.UpdateMetric(&metric)

		require.Error(t, err)
	})

	t.Run("concurrently ok", func(t *testing.T) {
		s, close := memstorage(t, 3*time.Minute)
		defer close()

		var wg sync.WaitGroup
		for range 10 {
			wg.Add(1)
			go func() {
				_, _ = s.UpdateMetric(&mCounter)
				_, _ = s.UpdateMetric(&mGauge)
				wg.Done()
			}()
		}
		got, err := s.UpdateMetric(&mGauge)

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
				s, close := memstorage(t, tc.interval)
				defer close()

				// Mock memstorage saver
				var saverCalled atomic.Bool
				s.saver = func(s *MemStorage) error { saverCalled.Store(true); return nil }

				// Update metric and give enough time to run goroutine
				_, err := s.UpdateMetric(&mCounter)
				require.NoError(t, err)
				time.Sleep(100 * time.Millisecond)

				assert.Equal(t, tc.expectCalled, saverCalled.Load())
			})
		}
	})
}

func TestMemStorage_CountMetric(t *testing.T) {
	s, deferFn := memstorage(t, 3*time.Minute)
	defer deferFn()
	_, _ = s.UpdateMetric(&models.Metric{Name: "foo", Type: models.CounterTypeName, Delta: 10})
	_, _ = s.UpdateMetric(&models.Metric{Name: "bar", Type: models.CounterTypeName, Delta: 200})
	_, _ = s.UpdateMetric(&models.Metric{Name: "goo", Type: models.GaugeTypeName, Value: 500.233})

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() { s.CountMetric(); wg.Done() }()
	}
	wg.Wait()
	got := s.CountMetric()

	assert.Equal(t, 3, got)
}

func TestMemStorage_GetMetric(t *testing.T) {
	fooCounter := models.Metric{Name: "foo", Type: models.CounterTypeName, Delta: 10}
	barCounter := models.Metric{Name: "bar", Type: models.CounterTypeName, Delta: 200}
	fooGauge := models.Metric{Name: "foo", Type: models.GaugeTypeName, Value: 500.233}

	s, deferFn := memstorage(t, 3*time.Minute)
	defer deferFn()
	_, _ = s.UpdateMetric(&fooCounter)
	_, _ = s.UpdateMetric(&barCounter)
	_, _ = s.UpdateMetric(&fooGauge)

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
				got, err := s.GetMetric(tc.args.mType, tc.args.mName)

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
				s.GetMetric(models.CounterTypeName, "foo") // nolint:errcheck
				s.GetMetric(models.GaugeTypeName, "foo")   // nolint:errcheck
				wg.Done()
			}()
		}
		wg.Wait()
		got, err := s.GetMetric(models.GaugeTypeName, "foo")

		assert.NoError(t, err)
		assert.EqualValues(t, fooGauge, got)
	})
}

func TestMemStorage_ListMetric(t *testing.T) {
	fooCounter := models.Metric{Name: "foo", Type: models.CounterTypeName, Delta: 10}
	barCounter := models.Metric{Name: "bar", Type: models.CounterTypeName, Delta: 200}
	fooGauge := models.Metric{Name: "foo", Type: models.GaugeTypeName, Value: 500.233}

	s, deferFn := memstorage(t, 3*time.Minute)
	defer deferFn()
	_, _ = s.UpdateMetric(&fooCounter)
	_, _ = s.UpdateMetric(&barCounter)
	_, _ = s.UpdateMetric(&fooGauge)

	t.Run("list ok", func(t *testing.T) {
		metrics, err := s.ListMetric()
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
				mResults[idx], _ = s.ListMetric()
				wg.Done()
			}()
		}

		wg.Wait()
		assert.Equal(t, 3, len(mResults[0]))
		assert.Equal(t, 3, len(mResults[1]))
		assert.Equal(t, 3, len(mResults[2]))
	})
}
