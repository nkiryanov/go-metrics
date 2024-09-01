package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/storage"
	"github.com/nkiryanov/go-metrics/internal/storage/mocks"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_UpdateMetricPlain(t *testing.T) {
	tests := []struct {
		name string

		storageUpdateErr error

		method       string
		url          string
		expectedCode int
		expectedBody string
	}{
		{
			"POST counter, ok",
			nil,
			http.MethodPost, "/update/counter/cpu-usage/100", http.StatusOK, "100",
		},
		{
			"POST gauge, ok",
			nil,
			http.MethodPost, "/update/gauge/cpu-usage/220.23", http.StatusOK, "220.23",
		},
		{
			"POST counter parse error, 400",
			nil,
			http.MethodPost, "/update/counter/cpu-usage/100.23", http.StatusBadRequest, "bad value to update counter\n",
		},
		{
			"POST gauge parse error, 400",
			nil,
			http.MethodPost, "/update/gauge/cpu-usage/some", http.StatusBadRequest, "bad value to update gauge\n",
		},
		{
			"GET metric, 405-NotAllowed",
			nil,
			http.MethodGet, "/update/counter/cpu-usage/100", http.StatusMethodNotAllowed, "",
		},
		{
			"POST metric storage err, 500",
			errors.New("oh no! storage failed :("),
			http.MethodPost, "/update/counter/cpu-usage/100", http.StatusInternalServerError, "oh no! storage failed :(\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockedStorage := &mocks.StorageMock{UpdateMetricFunc: func(m *models.Metric) (models.Metric, error) {
				var result models.Metric = *m
				return result, tc.storageUpdateErr
			}}

			router := NewMetricRouter(mockedStorage)
			srv := httptest.NewServer(router)
			defer srv.Close()

			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL + tc.url

			resp, err := req.Send()

			require.NoError(t, err)
			assert.Contains(t, resp.Header().Get("content-type"), "text/plain")
			assert.Equal(t, tc.expectedCode, resp.StatusCode())
			assert.Equal(t, tc.expectedBody, string(resp.Body()))
		})
	}
}

func TestHandlers_GetMetricPlain(t *testing.T) {
	cpuGauge := models.Metric{ID: "cpu-usage", MType: "gauge", Value: 23.23}
	emptyCounter := models.Metric{ID: "mem-usage", MType: "counter"}
	unknownMetric := models.Metric{ID: "mem-usage", MType: "unknown-type"}

	tests := []struct {
		name string

		storReturnValue models.Metric
		storReturnOk    bool
		storReturnErr   error

		method string
		url    string

		expectedCode int
		expectedBody string
	}{
		{
			"GET existed, ok",
			cpuGauge, true, nil,
			http.MethodGet, "/value/gauge/cpu-usage",
			http.StatusOK, "23.23",
		},
		{
			"GET not existed, 404",
			emptyCounter, false, nil,
			http.MethodGet, "/value/counter/mem-usage",
			http.StatusNotFound, "metric not found. type: counter, id: mem-usage\n",
		},
		{
			"GET unknown type, 404",
			unknownMetric, false, errors.New("storage error: unknown metric type"),
			http.MethodGet, "/value/unknown-type/mem-usage",
			http.StatusNotFound, "storage error: unknown metric type\n",
		},
		{
			"GET invalid url pattern, 404",
			unknownMetric, true, nil,
			http.MethodGet, "/value/co",
			http.StatusNotFound, "404 page not found\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockedStorage := &mocks.StorageMock{GetMetricFunc: func(mID string, mType string) (models.Metric, bool, error) {
				return tc.storReturnValue, tc.storReturnOk, tc.storReturnErr
			}}

			router := NewMetricRouter(mockedStorage)
			srv := httptest.NewServer(router)
			defer srv.Close()

			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL + tc.url

			resp, err := req.Send()

			require.NoError(t, err)
			assert.Equal(t, tc.expectedCode, resp.StatusCode())
			assert.Equal(t, tc.expectedBody, string(resp.Body()))
		})
	}
}

func TestHandlers_ListMetrics(t *testing.T) {
	fooCounter := &models.Metric{ID: "foo", MType: "counter", Delta: 100}
	barCounter := &models.Metric{ID: "bar", MType: "counter", Delta: 200}
	memGauge := &models.Metric{ID: "mem-load", MType: "gauge", Value: 234.23}

	stor := storage.NewMemStorage()
	_, _ = stor.UpdateMetric(fooCounter)
	_, _ = stor.UpdateMetric(barCounter)
	_, _ = stor.UpdateMetric(memGauge)

	router := NewMetricRouter(stor)
	srv := httptest.NewServer(router)
	defer srv.Close()

	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		expectedInBody []string
	}{
		{
			"GET list, ok",
			http.MethodGet, "/", http.StatusOK, []string{"foo", "bar", "mem-load", "100", "200", "234.23"},
		},
		{
			"POST list, 405-NotAllowed",
			http.MethodPost, "/", http.StatusMethodNotAllowed, []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = tc.method
			req.URL = srv.URL + tc.url

			resp, err := req.Send()

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode())
			for _, shouldInBody := range tc.expectedInBody {
				assert.Contains(t, string(resp.Body()), shouldInBody)
			}
		})
	}
}
