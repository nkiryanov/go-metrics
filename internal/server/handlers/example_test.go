package handlers_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/server/handlers"
	"github.com/nkiryanov/go-metrics/internal/server/storage/mocks"
)

func ExampleNewMetricRouter_updateCounter() {
	storage := &mocks.StorageMock{
		UpdateMetricFunc: func(ctx context.Context, m models.Metric) (models.Metric, error) {
			return m, nil
		},
	}

	router := handlers.NewMetricRouter(storage, logger.NewNoOpLogger(), "")
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/update/counter/requests/100", "text/plain", nil)
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode, string(body))

	// Output:
	// 200 100
}

func ExampleNewMetricRouter_updateGauge() {
	storage := &mocks.StorageMock{
		UpdateMetricFunc: func(ctx context.Context, m models.Metric) (models.Metric, error) {
			return m, nil
		},
	}

	router := handlers.NewMetricRouter(storage, logger.NewNoOpLogger(), "")
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/update/gauge/cpu/45.67", "text/plain", nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode, string(body))

	// Output:
	// 200 45.67
}

func ExampleNewMetricRouter_updateJSON() {
	storage := &mocks.StorageMock{
		UpdateMetricFunc: func(ctx context.Context, m models.Metric) (models.Metric, error) {
			return m, nil
		},
	}

	router := handlers.NewMetricRouter(storage, logger.NewNoOpLogger(), "")
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/update/", "application/json", strings.NewReader(`{"id":"req","type":"counter","delta":42}`))
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode, string(body))

	// Output:
	// 200 {"id":"req","type":"counter","delta":42}
}

func ExampleNewMetricRouter_getValue() {
	storage := &mocks.StorageMock{
		GetMetricFunc: func(ctx context.Context, mType, mName string) (models.Metric, error) {
			return models.Metric{Type: models.GaugeTypeName, Name: "cpu", Value: 78.5}, nil
		},
	}

	router := handlers.NewMetricRouter(storage, logger.NewNoOpLogger(), "")
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/value/gauge/cpu")
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode, string(body))

	// Output:
	// 200 78.5
}

func ExampleNewMetricRouter_ping() {
	storage := &mocks.StorageMock{
		PingFunc: func(ctx context.Context) error { return nil },
	}

	router := handlers.NewMetricRouter(storage, logger.NewNoOpLogger(), "")
	srv := httptest.NewServer(router)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ping")
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.StatusCode, string(body))

	// Output:
	// 200 OK
}
