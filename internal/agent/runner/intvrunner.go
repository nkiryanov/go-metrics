package runner

import (
	"context"
	"time"
)

// Simple interval runner
// Executes a function at regular intervals until it is stopped
// There is no graceful shutdown, may be implemented later
type IntvRunner struct {
	// Interval at which the task is executed
	intv time.Duration

	// Start delay
	delay time.Duration
}

func NewIntvRunner(d time.Duration, intv time.Duration) *IntvRunner {
	return &IntvRunner{
		intv:  intv,
		delay: d,
	}
}

func (ir IntvRunner) Run(ctx context.Context, fn func()) {
	gofn := func() {
		fn()

		ticker := time.NewTicker(ir.intv)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fn()
				ticker.Reset(ir.intv)
			}
		}
	}

	// Dummy stop, replaced real one if delay set
	stop := func() bool { return true }

	// Wait to start runner
	if ir.delay > 0 {
		timer := time.AfterFunc(ir.delay, gofn)
		stop = timer.Stop
	} else {
		go gofn()
	}

	// Wait for signal to stop
	<-ctx.Done()

	// Stop delayed start if canceled before the delay elapses
	stop()
}
