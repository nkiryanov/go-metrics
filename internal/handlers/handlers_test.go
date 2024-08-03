package handlers

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/nkiryanov/go-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func prefillCounter(s storage.Storage, name string, value string) {
	cv, _ := strconv.ParseInt(value, 10, 64)
	s.UpdateCounter(storage.MetricName(name), storage.Countable(cv))
}

func prefillGauge(s storage.Storage, name string, value string) {
	gv, _ := strconv.ParseFloat(value, 64)
	s.SetGauge(storage.MetricName(name), storage.Gaugeable(gv))
}

func register(pattern string, h http.HandlerFunc) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, h)
	return mux
}

func TestMetricsAPI_UpdateCounter(t *testing.T) {
	testCases := []struct {
		name                string
		metricName          string
		metricValue         string
		expectedCode        int
		expectedIsStored    bool
		expectedStoredValue storage.Countable
	}{
		{
			name:                "update with valid counter",
			metricName:          "cpu-usage",
			metricValue:         "10",
			expectedCode:        http.StatusOK,
			expectedIsStored:    true,
			expectedStoredValue: 10,
		},
		{
			name:                "update with flout should fail",
			metricName:          "cpu-usage",
			metricValue:         "10.23",
			expectedCode:        http.StatusBadRequest,
			expectedIsStored:    false,
			expectedStoredValue: 0,
		},
		{
			name:                "update with string should fail",
			metricName:          "cpu-usage",
			metricValue:         "invalid-value",
			expectedCode:        http.StatusBadRequest,
			expectedIsStored:    false,
			expectedStoredValue: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			api := NewMetricsAPI(storage.NewMemStorage())
			mux := register("/update/counter/{mName}/{mValue}", api.UpdateCounter)
			r := httptest.NewRequest(http.MethodPost, "/update/counter/"+tc.metricName+"/"+tc.metricValue, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedCode, w.Code)
			vStored, isInStore := api.storage.GetCounter(storage.MetricName(tc.metricName))
			assert.Equal(t, tc.expectedIsStored, isInStore)
			assert.EqualValues(t, tc.expectedStoredValue, vStored)
		})
	}
}

func TestMetricsAPI_UpdateGauge(t *testing.T) {
	testCases := []struct {
		name                string
		metricName          string
		metricValue         string
		expectedCode        int
		expectedIsStored    bool
		expectedStoredValue storage.Gaugeable
	}{
		{
			name:                "update with valid counter",
			metricName:          "memory",
			metricValue:         "10.23",
			expectedCode:        http.StatusOK,
			expectedIsStored:    true,
			expectedStoredValue: 10.23,
		},
		{
			name:                "update with string should fail",
			metricName:          "memory",
			metricValue:         "invalid",
			expectedCode:        http.StatusBadRequest,
			expectedIsStored:    false,
			expectedStoredValue: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			api := NewMetricsAPI(storage.NewMemStorage())
			mux := register("/update/gauge/{mName}/{mValue}", api.UpdateGauge)
			r := httptest.NewRequest(http.MethodPost, "/update/gauge/"+tc.metricName+"/"+tc.metricValue, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedCode, w.Code)
			vStored, isInStore := api.storage.GetGauge(storage.MetricName(tc.metricName))
			assert.Equal(t, tc.expectedIsStored, isInStore)
			assert.EqualValues(t, tc.expectedStoredValue, vStored)
		})
	}
}

func TestMetricsAPI_UpdateCounterWithPrefilledStorage(t *testing.T) {
	api := NewMetricsAPI(storage.NewMemStorage())
	prefillCounter(api.storage, "cpu-usage", "10")
	mux := register("/update/counter/{mName}/{mValue}", api.UpdateCounter)
	r := httptest.NewRequest(http.MethodPost, "/update/counter/cpu-usage/20", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	vStored, isInStore := api.storage.GetCounter(storage.MetricName("cpu-usage"))
	assert.True(t, isInStore)
	assert.EqualValues(t, 30, vStored, "stored value should be sum of prefill and new value")
}

func TestMetricsAPI_UpdateGaugeWithPrefilledStorage(t *testing.T) {
	api := NewMetricsAPI(storage.NewMemStorage())
	prefillGauge(api.storage, "memory", "10.23")
	mux := register("/update/gauge/{mName}/{mValue}", api.UpdateGauge)
	r := httptest.NewRequest(http.MethodPost, "/update/gauge/memory/20.23", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	vStored, isInStore := api.storage.GetGauge(storage.MetricName("memory"))
	assert.True(t, isInStore)
	assert.EqualValues(t, 20.23, vStored, "stored value should be replaced with new value")
}
