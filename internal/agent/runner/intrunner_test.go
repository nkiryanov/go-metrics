package runner

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func countFn() (w func(), r func() int) {
	var mu sync.Mutex
	called := 0

	w = func() {
		mu.Lock()
		called += 1
		mu.Unlock()
	}
	r = func() int {
		mu.Lock()
		defer mu.Unlock()
		return called
	}
	return
}

func TestRunner_IntvRunner(t *testing.T) {
	t.Run("stopped by context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		fn, reader := countFn()

		r := NewIntvRunner(0, 100*time.Millisecond)
		r.Run(ctx, fn)

		assert.LessOrEqual(t, 3, reader(), "looks like passed fn not executed")
	})

	t.Run("started with delay", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		defer cancel()
		fn, reader := countFn()

		r := NewIntvRunner(250*time.Millisecond, 100*time.Millisecond)
		r.Run(ctx, fn)

		assert.Equal(t, 1, reader(), "should be executed once and cancelled")
	})

	t.Run("cancel before delay ok", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		fn, reader := countFn()

		r := NewIntvRunner(150*time.Millisecond, 10*time.Millisecond)
		r.Run(ctx, fn)

		time.Sleep(100 * time.Millisecond) // Give enough time to start fn if not cancelled correctly
		assert.Equal(t, 0, reader(), "should NOT be executed cause cancelled before delay passed")
	})

	t.Run("executed with interval", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
		defer cancel()
		fn, reader := countFn()

		r := NewIntvRunner(5*time.Millisecond, 10*time.Millisecond)
		r.Run(ctx, fn)

		assert.Equal(t, 3, reader(), "should be executed at: 5, 15, 25 milliseconds")
	})
}
