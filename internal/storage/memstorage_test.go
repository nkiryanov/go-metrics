package storage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Some storable that not supported by storage
type yastorable string

func (ys yastorable) String() string { return "not-supported-storable" }
func (ya yastorable) Type() string   { return "no-supported" }

func updateConcurrently(s *MemStorage, key string, value Storable, count int) {
	var wg sync.WaitGroup

	for range count {
		wg.Add(2)
		go func() {
			s.UpdateMetric(key, value) // nolint: errcheck
			wg.Done()
		}()

		go func() {
			switch value.(type) {
			case Counter:
				s.GetCounter(key)
			case Gauge:
				s.GetGauge(key)
			default:
				panic("what are you doing bro?!")
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestMemStorage_GetCounter(t *testing.T) {
	type expectedResult struct {
		value Counter
		ok    bool
	}
	tCases := []struct {
		name     string
		mName    string
		expected expectedResult
	}{
		{
			name:     "not existed counter",
			mName:    "not-existed",
			expected: expectedResult{value: 0, ok: false},
		},
		{
			name:     "existed counter",
			mName:    "foo",
			expected: expectedResult{value: 11, ok: true},
		},
	}

	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMemStorage()
			// prefill storage with counter 'foo'
			storage.UpdateCounter("foo", 11)

			value, ok := storage.GetCounter(tc.mName)

			assert.EqualValues(t, tc.expected, expectedResult{value, ok})
		})
	}
}

func TestMemStorage_UpdateCountable(t *testing.T) {
	storage := NewMemStorage()

	firstUpdated := storage.UpdateCounter("foo", 230)
	secondUpdated := storage.UpdateCounter("foo", 500)
	stored, ok := storage.GetCounter("foo")

	assert.EqualValues(t, 230, firstUpdated)
	assert.EqualValues(t, 230+500, secondUpdated, "existed countable should updated with value")
	assert.Equal(t, true, ok)
	assert.EqualValues(t, 230+500, stored, "stored valued should be sum of all updated")
}

func TestMemStorage_UpdateCounterConcurrently(t *testing.T) {
	storage := NewMemStorage()

	updateConcurrently(storage, "foo", Counter(1), 100)

	value, _ := storage.GetCounter("foo")
	assert.EqualValues(t, 100, value)
}

func TestMemStorage_IterateCounters(t *testing.T) {
	storage := NewMemStorage()
	storage.UpdateCounter("foo", 1)
	storage.UpdateCounter("bar", 2)
	var wg sync.WaitGroup

	// Do some concurrent access
	for range 10 {
		wg.Add(1)
		go func() {
			storage.IterateCounters(func(string, Counter) {})
			wg.Done()
		}()
	}

	got := func() Counter {
		var sum Counter
		storage.IterateCounters(func(_ string, value Counter) { sum += value })
		return sum
	}()

	assert.EqualValues(t, 3, got)
}

func TestMemStorage_GetGauge(t *testing.T) {
	type expectedResult struct {
		value Gauge
		ok    bool
	}
	tCases := []struct {
		name     string
		mName    string
		expected expectedResult
	}{
		{
			name:     "not existed gauge",
			mName:    "not-existed",
			expected: expectedResult{value: 0.0, ok: false},
		},
		{
			name:     "existed gauge",
			mName:    "foo",
			expected: expectedResult{value: 11.23, ok: true},
		},
	}

	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMemStorage()
			// prefill storage with gauge 'foo'
			storage.UpdateGauge("foo", 11.23)

			value, ok := storage.GetGauge(tc.mName)

			assert.EqualValues(t, tc.expected, expectedResult{value, ok})
		})
	}
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	storage := NewMemStorage()

	firstUpdated := storage.UpdateGauge("foo", 230.23)
	secondUpdated := storage.UpdateGauge("foo", 100.23)
	stored, _ := storage.GetGauge("foo")

	assert.EqualValues(t, 230.23, firstUpdated)
	assert.EqualValues(t, 100.23, secondUpdated, "updating gauge should replace stored value")
	assert.EqualValues(t, 100.23, stored, "stored value should match last updated call")
}

func TestMemStorage_UpdateGaugeConcurrently(t *testing.T) {
	storage := NewMemStorage()

	updateConcurrently(storage, "bar", Gauge(1.1), 100)

	value, _ := storage.GetGauge("bar")
	assert.EqualValues(t, 1.1, value)
}

func TestMemStorage_IterateGauges(t *testing.T) {
	storage := NewMemStorage()
	storage.UpdateGauge("foo", 1.1)
	storage.UpdateGauge("bar", 2.2)
	var wg sync.WaitGroup

	// Do some concurrent access
	for range 10 {
		wg.Add(1)
		go func() {
			storage.IterateGauges(func(string, Gauge) {})
			wg.Done()
		}()
	}

	got := func() Gauge {
		var sum Gauge
		storage.IterateGauges(func(_ string, value Gauge) { sum += value })
		return sum
	}()

	assert.InDelta(t, 3.3, float64(got), 1.e-7)
}

func TestMemStorage_Len(t *testing.T) {
	storage := NewMemStorage()
	storage.UpdateMetric("foo", Counter(10))    // nolint: errcheck
	storage.UpdateMetric("bar", Counter(200))   // nolint: errcheck
	storage.UpdateMetric("goo", Gauge(500.233)) // nolint: errcheck
	var wg sync.WaitGroup

	for range 10 {
		wg.Add(1)
		go func() { storage.Len(); wg.Done() }()
	}

	got := storage.Len()
	assert.Equal(t, 3, got)
}

func TestMemStorage_UpdateMetric(t *testing.T) {
	type fnArgs struct {
		mName  string
		mValue Storable
	}
	type expectedResult struct {
		expectedGot Storable
		ok          bool
	}
	tCases := []struct {
		name     string
		fnArgs   fnArgs
		expected expectedResult
	}{
		{
			name:     "update counter, ok",
			fnArgs:   fnArgs{"foo", Counter(10)},
			expected: expectedResult{expectedGot: Counter(10), ok: true},
		},
		{
			name:     "update gauge, bad",
			fnArgs:   fnArgs{"cpu-usage", Gauge(10.23)},
			expected: expectedResult{expectedGot: Gauge(10.23), ok: true},
		},
		{
			name:     "update not-supported-type, bad",
			fnArgs:   fnArgs{"not-supported-type", yastorable("its-storable-but-storage-not-support-it")},
			expected: expectedResult{expectedGot: nil, ok: false},
		},
	}

	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMemStorage()

			got, err := storage.UpdateMetric(tc.fnArgs.mName, tc.fnArgs.mValue)

			if tc.expected.ok {
				require.Nil(t, err)
			}
			if !tc.expected.ok {
				require.Error(t, err)
			}
			assert.EqualValues(t, tc.expected.expectedGot, got)
		})
	}

}

func TestMemStorage_Iterate(t *testing.T) {
	storage := NewMemStorage()
	storage.UpdateMetric("foo", Counter(10))    // nolint: errcheck
	storage.UpdateMetric("bar", Counter(200))   // nolint: errcheck
	storage.UpdateMetric("goo", Gauge(500.233)) // nolint: errcheck
	var wg sync.WaitGroup

	// Run something to imitate concurrent access
	for range 10 {
		wg.Add(1)
		go func() {
			storage.Iterate(func(mType string, mName string, mValue Storable) {})
			wg.Done()
		}()
	}

	// Iterate be all metrics and return slice of it
	metrics := func() []Storable {
		metrics := make([]Storable, 0)
		storage.Iterate(func(_ string, _ string, mValue Storable) { metrics = append(metrics, mValue) })
		return metrics
	}()

	require.Equal(t, 3, len(metrics))
	assert.Contains(t, metrics, Counter(10))
	assert.Contains(t, metrics, Counter(200))
	assert.Contains(t, metrics, Gauge(500.233))
}
