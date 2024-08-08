package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/nkiryanov/go-metrics/internal/storage"
	"github.com/nkiryanov/go-metrics/internal/storage/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_UpdateMetric(t *testing.T) {
	mockedStorable := &mocks.StorableMock{StringFunc: func() string {
		return "parsed ok"
	}}

	tests := []struct {
		name string

		parseError        error
		updateMetricError error

		method       string
		url          string
		expectedCode int
		expectedBody string
	}{
		{
			"Post metric, parsed and saved to store ok",
			nil, nil,
			http.MethodPost, "/update/counter/cpu-usage/100", http.StatusOK, "saved-to-store-ok",
		},
		{
			"Get metric, should 405",
			nil, nil,
			http.MethodGet, "/update/counter/cpu-usage/100", http.StatusMethodNotAllowed, "",
		},
		{
			"Post metric, but parsing error occur",
			errors.New("oh no! parsing error"), nil,
			http.MethodPost, "/update/counter/cpu-usage/100.23", http.StatusBadRequest, "oh no! parsing error\n",
		},
		{
			"Post metric, but storage save error occurred",
			nil, errors.New("oh no! storage failed :("),
			http.MethodPost, "/update/counter/cpu-usage/100", http.StatusInternalServerError, "oh no! storage failed :(\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockedParser := &mocks.StorableParserMock{ParseFunc: func(mType string, s string) (storage.Storable, error) {
				return mockedStorable, tc.parseError
			}}

			mockedStorage := &mocks.StorageMock{UpdateMetricFunc: func(mName string, mValue storage.Storable) (storage.Storable, error) {
				return &mocks.StorableMock{StringFunc: func() string { return "saved-to-store-ok" }}, tc.updateMetricError
			}}

			router := NewMetricRouter(mockedStorage, mockedParser)
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

func TestHandlers_GetMetric(t *testing.T) {
	tests := []struct {
		name string

		storReturnValue storage.Storable
		storReturnOk    bool
		storReturnErr   error

		method string
		url    string

		expectedCode int
		expectedBody string
	}{
		{
			"Get-existed-metric",
			&mocks.StorableMock{StringFunc: func() string { return "100" }}, true, nil,
			http.MethodGet, "/value/counter/cpu-usage",
			http.StatusOK, "100",
		},
		{
			"NotFound if metric valid but not found",
			nil, false, nil,
			http.MethodGet, "/value/counter/mem-usage",
			http.StatusNotFound, "metric not found. type: counter, name: mem-usage\n",
		},
		{
			"NotFound if metric type unknown",
			nil, false, errors.New("storage error: unknown metric type"),
			http.MethodGet, "/value/unknown-type/mem-usage",
			http.StatusNotFound, "storage error: unknown metric type\n",
		},
		{
			"NotFound if url not valid",
			nil, true, nil,
			http.MethodGet, "/value/co",
			http.StatusNotFound, "404 page not found\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockedStorage := &mocks.StorageMock{GetMetricFunc: func(mt string, mv string) (storage.Storable, bool, error) {
				return tc.storReturnValue, tc.storReturnOk, tc.storReturnErr
			}}

			mockedParser := &mocks.StorableParserMock{}

			router := NewMetricRouter(mockedStorage, mockedParser)
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
	stor := storage.NewMemStorage()
	stor.UpdateCounter("foo", 100)
	stor.UpdateCounter("bar", 200)
	stor.UpdateGauge("mem-load", 234.23)

	router := NewMetricRouter(stor, storage.MemParser{})
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
			"List metrics",
			http.MethodGet, "/", http.StatusOK, []string{"foo", "bar", "mem-load", "100", "200", "234.23"},
		},
		{
			"List post method should not allowed",
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
