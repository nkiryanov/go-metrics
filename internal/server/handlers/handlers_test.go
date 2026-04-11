package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/server/storage"
	"github.com/nkiryanov/go-metrics/internal/server/storage/mocks"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_UpdateMetricPlain(t *testing.T) {
	tests := []struct {
		name string

		storageUpdateErr error

		method              string
		url                 string
		expectedCode        int
		expectedContentType string
		expectedBody        string
	}{
		{
			"POST counter, ok",
			nil,
			http.MethodPost, "/update/counter/cpu-usage/100", http.StatusOK, "text/plain", "100",
		},
		{
			"POST gauge, ok",
			nil,
			http.MethodPost, "/update/gauge/cpu-usage/220.23", http.StatusOK, "text/plain", "220.23",
		},
		{
			"POST counter parse error, 400",
			nil,
			http.MethodPost, "/update/counter/cpu-usage/100.23", http.StatusBadRequest, "text/plain", "bad value to update counter\n",
		},
		{
			"POST gauge parse error, 400",
			nil,
			http.MethodPost, "/update/gauge/cpu-usage/some", http.StatusBadRequest, "text/plain", "bad value to update gauge\n",
		},
		{
			"GET metric, 405-NotAllowed",
			nil,
			http.MethodGet, "/update/counter/cpu-usage/100", http.StatusMethodNotAllowed, "", "",
		},
		{
			"POST metric storage err, 500",
			errors.New("oh no! storage failed :("),
			http.MethodPost, "/update/counter/cpu-usage/100", http.StatusInternalServerError, "text/plain", "oh no! storage failed :(\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockedStorage := &mocks.StorageMock{UpdateMetricFunc: func(ctx context.Context, m models.Metric) (models.Metric, error) {
				var result = m
				return result, tc.storageUpdateErr
			}}

			router := NewMetricRouter(mockedStorage, logger.NewNoOpLogger(), "", nil)
			srv := httptest.NewServer(router)
			defer srv.Close()

			req := resty.New().R().SetHeader("Accept-Encoding", "")
			req.Method = tc.method
			req.URL = srv.URL + tc.url

			resp, err := req.Send()

			require.NoError(t, err)
			assert.Contains(t, resp.Header().Get("content-type"), tc.expectedContentType)
			assert.Equal(t, tc.expectedCode, resp.StatusCode())
			assert.Equal(t, tc.expectedBody, string(resp.Body()))
		})
	}
}

func TestHandlers_UpdateMetricJSON(t *testing.T) {
	t.Run("POST ok", func(t *testing.T) {
		mockedStorage := &mocks.StorageMock{UpdateMetricFunc: func(ctx context.Context, m models.Metric) (models.Metric, error) {
			return m, nil
		}}
		router := NewMetricRouter(mockedStorage, logger.NewNoOpLogger(), "", nil)
		srv := httptest.NewServer(router)
		defer srv.Close()

		resp, err := resty.New().
			R().
			SetHeader("Accept-Encoding", "").
			SetBody(`{"id": "cpu-usage", "type": "counter", "delta": 100}`).
			Post(srv.URL + "/update/")

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())
		assert.Equal(t, "application/json", resp.Header().Get("content-type"))
		assert.JSONEq(t, `{"id":"cpu-usage","type":"counter","delta":100}`, resp.String())
	})
}

func TestHandlers_GetMetricPlain(t *testing.T) {
	cpuGauge := models.Metric{Name: "cpu-usage", Type: "gauge", Value: 23.23}
	emptyCounter := models.Metric{Name: "mem-usage", Type: "counter"}
	unknownMetric := models.Metric{Name: "mem-usage", Type: "unknown-type"}

	tests := []struct {
		name string

		storReturnValue models.Metric
		storReturnErr   error

		method string
		url    string

		expectedCode int
		expectedBody string
	}{
		{
			"GET existed, ok",
			cpuGauge, nil,
			http.MethodGet, "/value/gauge/cpu-usage",
			http.StatusOK, "23.23",
		},
		{
			"GET not existed, 404",
			emptyCounter, storage.ErrNoMetric,
			http.MethodGet, "/value/counter/mem-usage",
			http.StatusNotFound, "metric not found. type: counter, id: mem-usage",
		},
		{
			"GET unknown type, 404",
			unknownMetric, storage.ErrNoMetric,
			http.MethodGet, "/value/unknown-type/mem-usage",
			http.StatusNotFound, "metric not found. type: unknown-type, id: mem-usage",
		},
		{
			"GET invalid url pattern, 404",
			unknownMetric, storage.ErrNoMetric,
			http.MethodGet, "/value/co",
			http.StatusNotFound, "404 page not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockedStorage := &mocks.StorageMock{GetMetricFunc: func(ctx context.Context, mType string, mName string) (models.Metric, error) {
				return tc.storReturnValue, tc.storReturnErr
			}}

			router := NewMetricRouter(mockedStorage, logger.NewNoOpLogger(), "", nil)
			srv := httptest.NewServer(router)
			defer srv.Close()

			req := resty.New().R().SetHeader("Accept-Encoding", "")
			req.Method = tc.method
			req.URL = srv.URL + tc.url

			resp, err := req.Send()

			require.NoError(t, err)
			assert.Equal(t, tc.expectedCode, resp.StatusCode())
			assert.Equal(t, tc.expectedBody, resp.String())
		})
	}
}

func TestHandlers_GetMetricJSON(t *testing.T) {
	cpuGauge := models.Metric{Name: "cpu-usage", Type: "gauge", Value: 23.23}
	emptyCounter := models.Metric{Name: "mem-usage", Type: "counter"}

	tests := []struct {
		name string

		storReturnValue models.Metric
		storReturnErr   error

		method  string
		request string

		expectedCode int
		expectedBody string
	}{
		{
			"GET existed, ok",
			cpuGauge, nil,
			http.MethodGet, `{"id": "cpu-usage", "type": "gauge"}}`,
			http.StatusOK, `{"id":"cpu-usage","type":"gauge","value":23.23}`,
		},
		{
			"GET not existed, 404",
			emptyCounter, storage.ErrNoMetric,
			http.MethodGet, `{"id": "mem-usage", "type": "counter"}}`,
			http.StatusNotFound, "metric not found. type: counter, id: mem-usage",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockedStorage := &mocks.StorageMock{GetMetricFunc: func(ctx context.Context, mType string, mName string) (models.Metric, error) {
				return tc.storReturnValue, tc.storReturnErr
			}}

			router := NewMetricRouter(mockedStorage, logger.NewNoOpLogger(), "", nil)
			srv := httptest.NewServer(router)
			defer srv.Close()

			resp, err := resty.New().
				R().
				SetHeader("Accept-Encoding", "").
				SetBody(tc.request).
				Post(srv.URL + "/value/")

			require.NoError(t, err)
			assert.Equal(t, tc.expectedCode, resp.StatusCode())
			assert.Equal(t, tc.expectedBody, resp.String())
		})
	}
}

func TestHandlers_ListMetrics(t *testing.T) {
	// Mocked storage; behave like it has 3 stored metrics.
	mockedStorage := &mocks.StorageMock{
		CountMetricFunc: func(ctx context.Context) (int, error) { return 3, nil },
		ListMetricFunc: func(ctx context.Context) ([]models.Metric, error) {
			return []models.Metric{
				{Type: models.CounterTypeName, Name: "bar", Delta: 200},
				{Type: models.CounterTypeName, Name: "foo", Delta: 100},
				{Type: models.GaugeTypeName, Name: "mem-load", Value: 234.23},
			}, nil
		},
	}

	router := NewMetricRouter(mockedStorage, logger.NewNoOpLogger(), "", nil)
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
			req.SetHeader("Accept-Encoding", "")
			req.Method = tc.method
			req.URL = srv.URL + tc.url

			resp, err := req.Send()

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode())
			for _, shouldInBody := range tc.expectedInBody {
				assert.Contains(t, resp.String(), shouldInBody)
			}
		})
	}
}

func TestHandler_Ping(t *testing.T) {
	t.Run("500 if ping fail", func(t *testing.T) {
		mockedStorage := &mocks.StorageMock{
			PingFunc: func(ctx context.Context) error { return errors.New("something terrible happened") },
		}

		router := NewMetricRouter(mockedStorage, logger.NewNoOpLogger(), "", nil)
		srv := httptest.NewServer(router)
		defer srv.Close()

		resp, err := resty.New().R().SetHeader("Accept-Encoding", "").Get(srv.URL + "/ping")

		require.NoError(t, err)
		assert.Equal(t, 500, resp.StatusCode())
	})

	t.Run("200 if ok", func(t *testing.T) {
		mockedStorage := &mocks.StorageMock{
			PingFunc: func(ctx context.Context) error { return nil },
		}

		router := NewMetricRouter(mockedStorage, logger.NewNoOpLogger(), "", nil)
		srv := httptest.NewServer(router)
		defer srv.Close()

		resp, err := resty.New().R().SetHeader("Accept-Encoding", "").Get(srv.URL + "/ping")

		require.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode())
	})
}
