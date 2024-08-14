package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/nkiryanov/go-metrics/internal/storage"

	"github.com/go-chi/chi/v5"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func registerRouter(s storage.Storage) chi.Router {
	api := NewMetricsAPI(s, nil)

	router := chi.NewRouter()
	router.Route("/", api.RegisterRoutes)

	return router
}

func prefillCounter(s storage.Storage, name string, value string) {
	cv, _ := strconv.ParseInt(value, 10, 64)
	s.UpdateCounter(storage.MetricName(name), storage.Countable(cv))
}

func prefillGauge(s storage.Storage, name string, value string) {
	gv, _ := strconv.ParseFloat(value, 64)
	s.SetGauge(storage.MetricName(name), storage.Gaugeable(gv))
}

func prepareRequest(method string, url string) (w *httptest.ResponseRecorder, r *http.Request) {
	r = httptest.NewRequest(method, url, nil)
	w = httptest.NewRecorder()
	return
}

func TestMetricAPI_UpdateCounterResponse(t *testing.T) {
	storage := storage.NewMemStorage()
	prefillCounter(storage, "cpu-usage", "100")
	router := registerRouter(storage)
	w, r := prepareRequest("POST", "/update/counter/cpu-usage/100")

	router.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "200")
}

func TestMetricsAPI_UpdateCounterStored(t *testing.T) {
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
			s := storage.NewMemStorage()
			router := registerRouter(s)
			w, r := prepareRequest("POST", fmt.Sprintf("/update/counter/%s/%s", tc.metricName, tc.metricValue))

			router.ServeHTTP(w, r)

			require.Equal(t, tc.expectedCode, w.Code)
			actuallyStored, isInStore := s.GetCounter(storage.MetricName(tc.metricName))
			assert.Equal(t, tc.expectedIsStored, isInStore)
			assert.EqualValues(t, tc.expectedStoredValue, actuallyStored)
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
			s := storage.NewMemStorage()
			router := registerRouter(s)
			w, r := prepareRequest("POST", fmt.Sprintf("/update/gauge/%s/%s", tc.metricName, tc.metricValue))

			router.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedCode, w.Code)
			actuallyStored, isInStore := s.GetGauge(storage.MetricName(tc.metricName))
			assert.Equal(t, tc.expectedIsStored, isInStore)
			assert.EqualValues(t, tc.expectedStoredValue, actuallyStored)
		})
	}
}

func TestMetricsAPI_UpdateInvalidMetricType(t *testing.T) {
	s := storage.NewMemStorage()
	router := registerRouter(s)
	w, r := prepareRequest("POST", "/update/invalid/metric/10")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMetricsAPI_UpdateCounterWithPrefilledStorage(t *testing.T) {
	s := storage.NewMemStorage()
	router := registerRouter(s)
	w, r := prepareRequest("POST", "/update/counter/poll-count/20")
	// Set 'poll-count' counter metric in storage
	prefillCounter(s, "poll-count", "10")

	router.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)
	actuallyStored, isInStore := s.GetCounter(storage.MetricName("poll-count"))
	assert.True(t, isInStore)
	assert.EqualValues(t, 30, actuallyStored, "stored value should be sum of prefill and new value")
}

func TestMetricsAPI_UpdateGaugeWithPrefilledStorage(t *testing.T) {
	s := storage.NewMemStorage()
	router := registerRouter(s)
	w, r := prepareRequest("POST", "/update/gauge/memory/20.23")
	// Set 'memory' gauge metric in storage
	prefillGauge(s, "memory", "10.23")

	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	vStored, isInStore := s.GetGauge(storage.MetricName("memory"))
	assert.True(t, isInStore)
	assert.EqualValues(t, 20.23, vStored, "stored value should be replaced with new value")
}

func TestMetricsApi_getMetric(t *testing.T) {
	s := storage.NewMemStorage()
	prefillCounter(s, "poll-count", "10")
	prefillGauge(s, "memory", "10.23")

	testCases := []struct {
		name          string
		path          string
		expectedCode  int
		expectedValue string
	}{
		{
			name:          "get counter metric",
			path:          "/value/counter/poll-count",
			expectedCode:  http.StatusOK,
			expectedValue: "10",
		},
		{
			name:          "get gauge metric",
			path:          "/value/gauge/memory",
			expectedCode:  http.StatusOK,
			expectedValue: "10.23",
		},
		{
			name:          "error unknown metric",
			path:          "/value/counter/unknown",
			expectedCode:  http.StatusNotFound,
			expectedValue: "Metric not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := registerRouter(s)
			w, r := prepareRequest(http.MethodGet, tc.path)

			router.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedCode, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedValue)
		})
	}
}
