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

func TestHandler_parse(t *testing.T) {
	type expected struct {
		value   storage.Storable
		isError bool
	}
	type fnArgs struct {
		mType string
		value string
	}
	tests := []struct {
		name     string
		fnArgs   fnArgs
		expected expected
	}{
		{
			name:     "valid counter",
			fnArgs:   fnArgs{mType: "counter", value: "20"},
			expected: expected{storage.Counter(20), false},
		},
		{
			name:     "invalid counter",
			fnArgs:   fnArgs{mType: "counter", value: "20.123"},
			expected: expected{storage.Counter(0), true},
		},
		{
			name:     "valid gauge",
			fnArgs:   fnArgs{mType: "gauge", value: "20.123"},
			expected: expected{storage.Gauge(20.123), false},
		},
		{
			name:     "invalid gauge",
			fnArgs:   fnArgs{mType: "gauge", value: "not-valid-gauge"},
			expected: expected{storage.Gauge(0.0), true},
		},
		{
			name:     "invalid metric type",
			fnArgs:   fnArgs{mType: "invalid-type", value: "10"},
			expected: expected{nil, true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parse(tc.fnArgs.mType, tc.fnArgs.value)

			if tc.expected.isError {
				require.Error(t, err)
			}
			assert.Equal(t, tc.expected.value, got)
		})
	}
}

func TestHandler_UpdateMetricPlain(t *testing.T) {
	tests := []struct {
		name string

		updateMetricError error

		method       string
		url          string
		expectedCode int
		expectedBody string
	}{
		{
			"POST metric, ok",
			nil,
			http.MethodPost, "/update/counter/cpu-usage/100", http.StatusOK, "saved-to-store-ok",
		},
		{
			"GET metric, 405-NotAllowed",
			nil,
			http.MethodGet, "/update/counter/cpu-usage/100", http.StatusMethodNotAllowed, "",
		},
		{
			"POST metric parse error, 400",
			nil,
			http.MethodPost, "/update/counter/cpu-usage/100.23", http.StatusBadRequest, "strconv.ParseInt: parsing \"100.23\": invalid syntax\n",
		},
		{
			"POST metric storage err, 500",
			errors.New("oh no! storage failed :("),
			http.MethodPost, "/update/counter/cpu-usage/100", http.StatusInternalServerError, "oh no! storage failed :(\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockedStorage := &mocks.StorageMock{UpdateMetricFunc: func(mName string, mValue storage.Storable) (storage.Storable, error) {
				return &mocks.StorableMock{StringFunc: func() string { return "saved-to-store-ok" }}, tc.updateMetricError
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
			"GET existed, ok",
			&mocks.StorableMock{StringFunc: func() string { return "100" }}, true, nil,
			http.MethodGet, "/value/counter/cpu-usage",
			http.StatusOK, "100",
		},
		{
			"GET not existed, 404",
			nil, false, nil,
			http.MethodGet, "/value/counter/mem-usage",
			http.StatusNotFound, "metric not found. type: counter, name: mem-usage\n",
		},
		{
			"GET unknown type, 404",
			nil, false, errors.New("storage error: unknown metric type"),
			http.MethodGet, "/value/unknown-type/mem-usage",
			http.StatusNotFound, "storage error: unknown metric type\n",
		},
		{
			"GET invalid url pattern, 404",
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
	stor := storage.NewMemStorage()
	stor.UpdateCounter("foo", 100)
	stor.UpdateCounter("bar", 200)
	stor.UpdateGauge("mem-load", 234.23)

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
