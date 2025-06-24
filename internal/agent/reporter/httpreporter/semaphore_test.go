package httpreporter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSemaphore(t *testing.T) {
	sem := NewSemaphore(2)
	started := make(chan struct{}, 3)
	done := make(chan struct{}, 3)

	for range 3 {
		go func() {
			sem.Acquire()
			started <- struct{}{}
			time.Sleep(50 * time.Millisecond)
			sem.Release()
			done <- struct{}{}
		}()
	}

	// after a short pause, at most 2 goroutines should start
	time.Sleep(10 * time.Millisecond)
	assert.Len(t, started, 2, "only 2 goroutines should have acquired the semaphore")

	// wait for all completions
	time.Sleep(100 * time.Millisecond)
	assert.Len(t, done, 3, "all 3 goroutines should have finished eventually")
}
