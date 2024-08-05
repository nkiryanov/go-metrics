package reporter

import (
	"context"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/storage"
	"github.com/stretchr/testify/require"
)

var thirdSecond = time.Millisecond * 300

func TestHTTPReporter_RunStopWithSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), thirdSecond)
	defer cancel()
	publisher, _ := NewHTTPReporter("http://example.com", time.Second, storage.NewMemStorage())

	err := publisher.Run(ctx)

	require.Equal(t, ErrReporterStopped, err)
}
