package publisher

import (
	"context"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var thirdSecond = time.Millisecond * 300

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

func TestHTTPPublisher_RunStopWithSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), thirdSecond)
	defer cancel()
	publisher, _ := NewHTTPPublisher("http://example.com", time.Second, storage.NewMemStorage())

	err := publisher.Run(ctx)

	require.Equal(t, ErrPublisherStopped, err)
}
