package storage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func updateCounterConcurrently(s *MemStorage, key MetricName, value Countable, count int) {
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

func setGaugeConcurrently(s *MemStorage, key MetricName, value Gaugeable, count int) {
	var wg sync.WaitGroup

	for range count {
		wg.Add(2)
		go func() {
			s.SetGauge(key, value)
			wg.Done()
		}()

		go func() {
			s.GetGauge(key)
			wg.Done()
		}()
	}

	wg.Wait()
}

func TestMemStorage_GetCounterNotExistedCounter(t *testing.T) {
	storage := NewMemStorage()

	value, ok := storage.GetCounter("foo")
	assert.False(t, ok)
	assert.Zero(t, value)
}

func TestMemStorage_GetGaugeNotExistedGauge(t *testing.T) {
	storage := NewMemStorage()

	value, ok := storage.GetGauge("bar")
	assert.False(t, ok)
	assert.Zero(t, value)
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	storage := NewMemStorage()

	storage.UpdateCounter("foo", 23)
	value, ok := storage.GetCounter("foo")

	assert.True(t, ok)
	assert.EqualValues(t, 23, value)
}

func TestMemStorage_UpdateCounterTwice(t *testing.T) {
	// Updating counter should sum all previous values
	storage := NewMemStorage()
	storage.UpdateCounter("foo", 99)

	storage.UpdateCounter("foo", 1)

	value, ok := storage.GetCounter("foo")
	assert.True(t, ok)
	assert.EqualValues(t, 100, value)
}

func TestMemStorage_SetGauge(t *testing.T) {
	storage := NewMemStorage()

	storage.SetGauge("bar", 23.23)
	value, ok := storage.GetGauge("bar")

	assert.True(t, ok)
	assert.EqualValues(t, 23.23, value)
}

func TestMemStorage_SetGaugeTwice(t *testing.T) {
	// Setting gauge should replace value
	storage := NewMemStorage()

	storage.SetGauge("bar", 23.23)
	storage.SetGauge("bar", 99.01)

	value, ok := storage.GetGauge("bar")
	assert.True(t, ok)
	assert.EqualValues(t, 99.01, value)
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
