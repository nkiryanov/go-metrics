package main

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/nkiryanov/go-metrics/internal/agent/collector"
	"github.com/nkiryanov/go-metrics/internal/agent/collector/memstatscollector"
	"github.com/nkiryanov/go-metrics/internal/agent/config"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter"
	"github.com/nkiryanov/go-metrics/internal/agent/reporter/httpreporter"
	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
)

type Agent struct {

	// Collectors to collect and access collected metrics
	Collectors []struct {
		n string              // Collector name
		c collector.Collector // Collector to run
		i time.Duration       // Collect interval
	}

	// Reporter requests metrics from collectors and report
	Reporter struct {
		n string            // Reporter name
		r reporter.Reporter // Reporter to run
		i time.Duration     // Report interval
	}

	Lgr logger.Logger
}

func NewAgent(cfg *config.Config) *Agent {
	lgr := logger.NewLogger(cfg.LogLevel)

	return &Agent{
		Collectors: []struct {
			n string
			c collector.Collector
			i time.Duration
		}{
			{n: "MemStats Collector", c: memstatscollector.New(), i: cfg.CollectInterval},
		},
		Reporter: struct {
			n string
			r reporter.Reporter
			i time.Duration
		}{
			n: "HTTP Reporter",
			r: httpreporter.New(
				cfg.ReportAddr,
				&http.Client{},
				[]time.Duration{time.Second, 3 * time.Second, 5 * time.Second},
				cfg.SecretKey,
				lgr,
			),
			i: cfg.ReportInterval,
		},
		Lgr: lgr,
	}
}

func (a *Agent) Run(ctx context.Context) {
	var wg sync.WaitGroup
	// Run collectors until context cancelled
	wg.Add(len(a.Collectors))
	for _, collr := range a.Collectors {
		go func() {
			defer wg.Done()
			a.Lgr.Info("Starting collector", "collector_name", collr.n, "collect_interval", collr.i)
			for ctx.Err() == nil {
				err := collector.Run(ctx, collr.c, collr.i)
				if err != nil {
					a.Lgr.Warn("Collector stopped with error", "error", err.Error())
				}
			}
		}()
	}

	// Run bath report with report interval, until context canceled
	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(a.Reporter.i)
		defer ticker.Stop()

		metrics := make([]models.Metric, 0)
		reportFn := func() error {
			metrics = metrics[:0]
			for _, collr := range a.Collectors {
				metrics = append(metrics, collr.c.List()...)
			}

			return a.Reporter.r.ReportBatch(metrics)
		}

		a.Lgr.Info("Starting reporter", "reporter_name", a.Reporter.n, "report_interval", a.Reporter.i)
		for ctx.Err() == nil {
			select {
			case <-ctx.Done():
				continue
			case <-ticker.C:
				err := reportFn()
				if err != nil {
					a.Lgr.Warn("Reporter batch report stopped with error", "error", err.Error())
				}
			}
		}
	}()

	wg.Wait()
}
