package main

import (
	"context"
	"testing"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/collector"
	cmocks "github.com/nkiryanov/go-metrics/internal/agent/collector/mocks"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter"
	rmocks "github.com/nkiryanov/go-metrics/internal/agent/reporter/mocks"
	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/stretchr/testify/require"
)

func TestAgent_Run(t *testing.T) {
	metric1 := models.Metric{Name: "first", Type: "counter", Delta: 3}
	metric2 := models.Metric{Name: "second", Type: "gauge", Value: 1.23}

	mockedReporter := &rmocks.ReporterMock{
		ReportBatchFunc: func(ms []models.Metric) error { return nil },
	}

	mockedCollector := &cmocks.CollectorMock{
		CollectFunc: func(_ context.Context) error { return nil },
		ListFunc:    func() []models.Metric { return []models.Metric{metric1} },
	}

	yaMockedCollector := &cmocks.CollectorMock{
		CollectFunc: func(_ context.Context) error { return nil },
		ListFunc:    func() []models.Metric { return []models.Metric{metric2} },
	}

	newAgent := func() *Agent {
		return &Agent{
			Collectors: []struct {
				n string              // Collector name
				c collector.Collector // Collector to run
				i time.Duration       // Collect interval
			}{
				{"collector1", mockedCollector, 100 * time.Millisecond},
				{"collector2", yaMockedCollector, 200 * time.Millisecond},
			},
			Reporter: struct {
				r reporter.Reporter
				i time.Duration
			}{
				r: mockedReporter, i: 300 * time.Millisecond,
			},
			Lgr: logger.NewNoOpLogger(),
		}
	}

	t.Run("stop on ctx cancel", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		t.Cleanup(cancel)
		a := newAgent()

		err := a.Run(ctx)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrAgentStopped)
	})

	t.Run("call reporter and collectors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		t.Cleanup(cancel)

		a := newAgent()

		err := a.Run(ctx)

		require.Error(t, err)
		require.ErrorIs(t, err, ErrAgentStopped)

		// Verify collectors were called
		require.GreaterOrEqual(t, len(mockedCollector.CollectCalls()), 3)   // 500 Millisecond is enough to call three times
		require.GreaterOrEqual(t, len(yaMockedCollector.CollectCalls()), 2) // 500 Millisecond is enough to call twice

		// Verify reporter was called
		require.GreaterOrEqual(t, len(mockedReporter.ReportBatchCalls()), 1)
		reportCallArgs := mockedReporter.ReportBatchCalls()[0]
		require.ElementsMatch(t, []models.Metric{metric1, metric2}, reportCallArgs.Metrics)
	})
}
