//go:build integration
// +build integration

package reporter

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var halfSecond = time.Millisecond * 500

type capturedReq struct {
	method string
	path   string
}

type testHTTPReceiver struct {
	requests []capturedReq
	lock     sync.Mutex
}

func newTestHTTPReceiver() testHTTPReceiver {
	return testHTTPReceiver{requests: make([]capturedReq, 0)}
}

func (rcv *testHTTPReceiver) Handler(w http.ResponseWriter, req *http.Request) {
	rcv.lock.Lock()
	rcv.requests = append(rcv.requests, capturedReq{req.Method, req.URL.Path})
	rcv.lock.Unlock()
}

func (rcv *testHTTPReceiver) run(listenAddr string, ctx context.Context) {
	// Run http receiver in background
	srv := &http.Server{
		Addr:    listenAddr,
		Handler: http.HandlerFunc(rcv.Handler),
	}

	go func() {
		srv.ListenAndServe() // nolint:errcheck
	}()

	go func() {
		<-ctx.Done()
		srv.Close()
	}()
}

func TestHTTPReporter_reportMetric(t *testing.T) {
	reporter, _ := NewHTTPReporter("http://localhost:51493", time.Millisecond, storage.NewMemStorage())
	// Run test http receiver
	ctx, stopRcv := context.WithTimeout(context.Background(), halfSecond)
	rcv := newTestHTTPReceiver()
	rcv.run("localhost:51493", ctx)

	status, err := reporter.reportMetric(storage.CounterTypeName, "counter-stat")

	stopRcv()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, rcv.requests, 1)
	assert.Equal(t, "POST", rcv.requests[0].method)
	assert.Equal(t, "/update/counter/counter-stat/0", rcv.requests[0].path)
}

func TestHTTPReporter_batchReport(t *testing.T) {
	// Prepare storage with some metrics
	s := storage.NewMemStorage()
	s.UpdateCounter("counter-stat", 1)
	s.UpdateGauge("gauge-stat", 2.01)
	// Prepare reporter
	reporter, _ := NewHTTPReporter("http://localhost:51493", time.Second, s)
	// Run test http receiver
	ctx, stopRcv := context.WithTimeout(context.Background(), halfSecond)
	rcv := newTestHTTPReceiver()
	rcv.run("localhost:51493", ctx)

	reporter.batchReport()

	stopRcv()
	require.Len(t, rcv.requests, 2)
	assert.Contains(t, rcv.requests, capturedReq{"POST", "/update/counter/counter-stat/1"})
	assert.Contains(t, rcv.requests, capturedReq{"POST", "/update/gauge/gauge-stat/2.01"})
}
