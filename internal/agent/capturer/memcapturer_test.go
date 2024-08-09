package capturer

import (
	"errors"
	"sync"
	"testing"

	"github.com/nkiryanov/go-metrics/internal/storage"
	"github.com/nkiryanov/go-metrics/internal/storage/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getStatNames(stats []Stat) []string {
	names := make([]string, 0, len(stats))
	for _, stat := range stats {
		names = append(names, stat.Name)
	}
	return names
}

func getCallNames(callInfo []struct {
	MName  string
	MValue storage.Storable
}) []string {
	names := make([]string, 0, len(callInfo))
	for _, info := range callInfo {
		names = append(names, info.MName)
	}
	return names
}

func TestMemCapturer_Capture(t *testing.T) {
	mc := NewMemCapturer()
	expectedNames := append(gauges, counters...)

	t.Run("capture return all stats", func(t *testing.T) {
		stats := mc.Capture()

		assert.EqualValues(t, expectedNames, getStatNames(stats))
	})

	t.Run("capture not race", func(t *testing.T) {
		var wg sync.WaitGroup

		// Run concurrently to possible race conditions
		for range 100 {
			wg.Add(1)
			go func() { mc.Capture(); wg.Done() }()
		}
		stats := mc.Capture()

		wg.Wait()
		assert.Equal(t, len(expectedNames), len(stats))
	})
}

func TestMemCapturer_CaptureWithSave(t *testing.T) {
	expectedNames := append(gauges, counters...)
	mc := NewMemCapturer()

	t.Run("update ok", func(t *testing.T) {
		mockedStorage := &mocks.StorageMock{UpdateMetricFunc: func(_ string, st storage.Storable) (storage.Storable, error) {
			return st, nil
		}}

		err := mc.CaptureWithSave(mockedStorage)

		require.NoError(t, err)
		callInfo := mockedStorage.UpdateMetricCalls()
		require.Equal(t, len(expectedNames), len(callInfo))
		assert.EqualValues(t, expectedNames, getCallNames(callInfo))
	})

	t.Run("return first err", func(t *testing.T) {
		// helper that fails only once on twenties (20) call
		failOnceFn := func() func(string, storage.Storable) (storage.Storable, error) {
			callCount := 0
			return func(_ string, st storage.Storable) (storage.Storable, error) {
				callCount += 1
				if callCount == 20 {
					return st, errors.New("just ask bill")
				}
				return st, nil
			}
		}

		mockedStorage := &mocks.StorageMock{UpdateMetricFunc: failOnceFn()}

		err := mc.CaptureWithSave(mockedStorage)

		require.Error(t, err)
		callInfo := mockedStorage.UpdateMetricCalls()
		require.Equal(t, len(expectedNames), len(callInfo))
		assert.EqualValues(t, expectedNames, getCallNames(callInfo))
	})
}
