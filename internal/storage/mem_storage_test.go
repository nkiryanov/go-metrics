package storage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func updateCounterConcurrently(s *MemStorage, key string, value Countable, count int) {
	var wg sync.WaitGroup

	for range count {
		wg.Add(2)
		go func() {
			s.UpdateCounter(key, value)
			wg.Done()
		}()

		go func() {
			s.GetCounter(key)
			wg.Done()
		}()
	}

	wg.Wait()
}

func setGaugeConcurrently(s *MemStorage, key string, value Gaugeable, count int) {
	var wg sync.WaitGroup

	for range count {
		wg.Add(2)
		go func() {
			s.UpdateGauge(key, value)
			wg.Done()
		}()

		go func() {
			s.GetGauge(key)
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestMemStorage_GetCounter(t *testing.T) {
	type expectedResult struct {
		value Countable
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

func TestMemStorage_GetGauge(t *testing.T) {
	type expectedResult struct {
		value Gaugeable
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

func TestMemStorage_UpdateCountable(t *testing.T) {
	storage := NewMemStorage()

	firstUpdated := storage.UpdateCounter("foo", 230)
	secondUpdated := storage.UpdateCounter("foo", 500)
	stored, ok := storage.GetCounter("foo")

	assert.Equal(t, 230, firstUpdated)
	assert.Equal(t, 230+500, secondUpdated, "existed countable should updated with value")
	assert.Equal(t, true, ok)
	assert.Equal(t, 230+500, stored, "stored valued should be sum of all updated")
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	storage := NewMemStorage()

	firstUpdated := storage.UpdateGauge("foo", 230.23)
	secondUpdated := storage.UpdateGauge("foo", 100.23)
	stored, _ := storage.GetGauge("foo")

	assert.Equal(t, 230.23, firstUpdated)
	assert.Equal(t, 100.23, secondUpdated, "updating gauge should replace stored value")
	assert.Equal(t, 100.23, stored, "stored value should match last updated call")
}

func TestMemStorage_UpdateCounterConcurrently(t *testing.T) {
	storage := NewMemStorage()

	updateCounterConcurrently(storage, "foo", 1, 100)

	value, _ := storage.GetCounter("foo")
	assert.EqualValues(t, 100, value)
}

func TestMemStorage_SetGaugeConcurrently(t *testing.T) {
	storage := NewMemStorage()

	setGaugeConcurrently(storage, "bar", 1.1, 100)

	value, _ := storage.GetGauge("bar")
	assert.EqualValues(t, 1.1, value)
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
			storage.IterateCounters(func(string, Countable) {})
			wg.Done()
		}()
	}

	got := func() Countable {
		var sum Countable
		storage.IterateCounters(func(_ string, value Countable) { sum += value })
		return sum
	}()

	assert.EqualValues(t, 3, got)
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
			storage.IterateGauges(func(string, Gaugeable) {})
			wg.Done()
		}()
	}

	got := func() Gaugeable {
		var sum Gaugeable
		storage.IterateGauges(func(_ string, value Gaugeable) { sum += value })
		return sum
	}()

	assert.InDelta(t, 3.3, float64(got), 1.e-7)
}

func TestMemStorage_Len(t *testing.T) {
	storage := NewMemStorage()
	storage.UpdateValue("counter", "foo", "10")
	storage.UpdateValue("counter", "bar", "200")
	storage.UpdateValue("gauge", "goo", "500.233")
	var wg sync.WaitGroup

	for range 10 {
		wg.Add(1)
		go func() { storage.Len(); wg.Done() }()
	}

	got := storage.Len()
	assert.Equal(t, 3, got)
}

func TestMemStorage_UpdateValue(t *testing.T) {
	type fnArgs struct {
		mType  string
		mName  string
		mValue string
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
			name:     "update counter with valid value",
			fnArgs:   fnArgs{"counter", "foo", "10"},
			expected: expectedResult{expectedGot: Countable(10), ok: true},
		},
		{
			name:     "update counter with invalid value",
			fnArgs:   fnArgs{"counter", "foo", "not-int-value"},
			expected: expectedResult{expectedGot: Countable(0), ok: false},
		},
		{
			name:     "update gauge with valid value",
			fnArgs:   fnArgs{"gauge", "gau", "10.23"},
			expected: expectedResult{expectedGot: Gaugeable(10.23), ok: true},
		},
		{
			name:     "update gauge with invalid data",
			fnArgs:   fnArgs{"gauge", "gua", "not-valid-value"},
			expected: expectedResult{expectedGot: Gaugeable(0.0), ok: false},
		},
		{
			name:     "update with incorrect metric at all",
			fnArgs:   fnArgs{"incorrect-type", "bar", "10"},
			expected: expectedResult{expectedGot: nil, ok: false},
		},
	}

	for _, tc := range tCases {
		t.Run(tc.name, func(t *testing.T) {
			storage := NewMemStorage()

			got, err := storage.UpdateValue(tc.fnArgs.mType, tc.fnArgs.mName, tc.fnArgs.mValue)

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
	storage.UpdateValue("counter", "foo", "10")
	storage.UpdateValue("counter", "bar", "200")
	storage.UpdateValue("gauge", "goo", "500.233")
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
	assert.Equal(t, Countable(10), metrics[0])
	assert.Equal(t, Countable(200), metrics[1])
	assert.Equal(t, Gaugeable(500.233), metrics[2])
}
