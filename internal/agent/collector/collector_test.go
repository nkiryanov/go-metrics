package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/collector/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCollector_Run(t *testing.T) {
	t.Run("stopped by context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 400*time.Millisecond)
		defer cancel()
		c := &mocks.CollectorMock{
			CollectFunc: func(ctx context.Context) error { return nil },
		}

		err := Run(ctx, c, 100*time.Millisecond)

		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.True(t, len(c.CollectCalls()) > 2, "looks like collect not called with timer")
	})

	t.Run("stop if collect fails", func(t *testing.T) {
		collectErr := errors.New("collect failed")
		var c *mocks.CollectorMock
		c = &mocks.CollectorMock{
			CollectFunc: func(ctx context.Context) error {
				// First few calls succeed, then fail
				if len(c.CollectCalls()) >= 3 {
					return collectErr
				}
				return nil
			},
		}

		err := Run(t.Context(), c, 50*time.Millisecond)

		assert.ErrorIs(t, err, collectErr)
		assert.GreaterOrEqual(t, len(c.CollectCalls()), 3, "should call collect at least 3 times before failing")
	})
}
