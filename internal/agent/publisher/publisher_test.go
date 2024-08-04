package publisher

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

const (
	halfSecond = time.Millisecond * 500
)

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

func TestPublisher_NewHTTPPublisherValidAddr(t *testing.T) {
	pubAddr := "http://example.com:8080"
	pubInterval := time.Second
	storage := storage.NewMemStorage()

	got, err := NewHTTPPublisher(pubAddr, pubInterval, storage)

	require.NoError(t, err)
	assert.Equal(t, "http://example.com:8080", got.pubAddr)
}

func TestPublisher_NewHTTPPublisherInvalidAddr(t *testing.T) {
	pubAddr := "localhost:8080"
	pubInterval := time.Second
	storage := storage.NewMemStorage()

	_, err := NewHTTPPublisher(pubAddr, pubInterval, storage)

	require.Error(t, err)
	assert.Equal(t, "publisher: Invalid publisher address", err.Error())
}

func TestHTTPPublisher_postMetric(t *testing.T) {
	publisher, _ := NewHTTPPublisher("http://localhost:51493", time.Millisecond, storage.NewMemStorage())
	// Run test http receiver
	ctx, stopRcv := context.WithTimeout(context.Background(), halfSecond)
	rcv := newTestHTTPReceiver()
	rcv.run("localhost:51493", ctx)

	status, err := publisher.postMetric(storage.CounterTypeName, "counter-stat")

	stopRcv()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, status)
	require.Len(t, rcv.requests, 1)
	assert.Equal(t, "POST", rcv.requests[0].method)
	assert.Equal(t, "/update/counter/counter-stat/0", rcv.requests[0].path)
}

func TestHTTPPublisher_batchPublish(t *testing.T) {
	// Prepare storage with some metrics
	s := storage.NewMemStorage()
	s.UpdateCounter("counter-stat", 1)
	s.SetGauge("gauge-stat", 2.01)
	// Prepare publisher
	publisher, _ := NewHTTPPublisher("http://localhost:51493", time.Second, s)
	// Run test http receiver
	ctx, stopRcv := context.WithTimeout(context.Background(), time.Second)
	rcv := newTestHTTPReceiver()
	rcv.run("localhost:51493", ctx)

	publisher.batchPublish()

	stopRcv()
	require.Len(t, rcv.requests, 2)
	assert.Contains(t, rcv.requests, capturedReq{"POST", "/update/counter/counter-stat/1"})
	assert.Contains(t, rcv.requests, capturedReq{"POST", "/update/gauge/gauge-stat/2.01"})
}

func TestHTTPPublisher_RunStopWithSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), halfSecond)
	defer cancel()
	publisher, _ := NewHTTPPublisher("http://example.com", time.Second, storage.NewMemStorage())

	err := publisher.Run(ctx)

	require.Equal(t, ErrPublisherStopped, err)
}
