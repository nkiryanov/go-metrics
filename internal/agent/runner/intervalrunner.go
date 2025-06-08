package runner

import (
	"context"
	"time"
)

// Simple interval runner
// Executes a function at regular intervals until it is stopped
// There is no graceful shutdown, may be implemented sometime
type IntervalRunner struct {
	// Interval at which the task is executed
	interval time.Duration

	// Start delay
	delay time.Duration
}

func NewIntvRunner(d time.Duration, interval time.Duration) *IntervalRunner {
	return &IntervalRunner{
		interval: interval,
		delay:    d,
	}
}

func (runner IntervalRunner) Run(ctx context.Context, fn func()) {
	gofn := func() {
		fn()

		ticker := time.NewTicker(runner.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fn()
				ticker.Reset(runner.interval)
			}
		}
	}

	// Dummy stop, will be replaced with real one if delay set
	stop := func() bool { return true }

	// Wait to start runner
	if runner.delay > 0 {
		timer := time.AfterFunc(runner.delay, gofn)
		stop = timer.Stop
	} else {
		go gofn()
	}

	// Wait for signal to stop
	<-ctx.Done()

	// Stop delayed start if canceled before the delay elapses
	stop()
}
