package poller

import (
	"context"
	"errors"
	"log/slog"
	"math/rand"
	"runtime"
	"time"

	"github.com/nkiryanov/go-metrics/internal/storage"
)

const (
	// Gauges captured by runtime.ReadMemStats
	Alloc         = storage.MetricName("Alloc")
	BuckHashSys   = storage.MetricName("BuckHashSys")
	Frees         = storage.MetricName("Frees")
	GCCPUFraction = storage.MetricName("GCCPUFraction")
	GCSys         = storage.MetricName("GCSys")
	HeapAlloc     = storage.MetricName("HeapAlloc")
	HeapIdle      = storage.MetricName("HeapIdle")
	HeapInuse     = storage.MetricName("HeapInuse")
	HeapObjects   = storage.MetricName("HeapObjects")
	HeapReleased  = storage.MetricName("HeapReleased")
	HeapSys       = storage.MetricName("HeapSys")
	LastGC        = storage.MetricName("LastGC")
	Lookups       = storage.MetricName("Lookups")
	MCacheInuse   = storage.MetricName("MCacheInuse")
	MCacheSys     = storage.MetricName("MCacheSys")
	MSpanInuse    = storage.MetricName("MSpanInuse")
	MSpanSys      = storage.MetricName("MSpanSys")
	Mallocs       = storage.MetricName("Mallocs")
	NextGC        = storage.MetricName("NextGC")
	NumForcedGC   = storage.MetricName("NumForcedGC")
	NumGC         = storage.MetricName("NumGC")
	OtherSys      = storage.MetricName("OtherSys")
	PauseTotalNs  = storage.MetricName("PauseTotalNs")
	StackInuse    = storage.MetricName("StackInuse")
	StackSys      = storage.MetricName("StackSys")
	Sys           = storage.MetricName("Sys")
	TotalAlloc    = storage.MetricName("TotalAlloc")
	RandomValue   = storage.MetricName("RandomValue")

	// Counters computed by the agent
	PollCount = storage.MetricName("PollCount")
)

var (
	gauges           = []storage.MetricName{Alloc, BuckHashSys, Frees, GCCPUFraction, GCSys, HeapAlloc, HeapIdle, HeapInuse, HeapObjects, HeapReleased, HeapSys, LastGC, Lookups, MCacheInuse, MCacheSys, MSpanInuse, MSpanSys, Mallocs, NextGC, NumForcedGC, NumGC, OtherSys, PauseTotalNs, StackInuse, StackSys, Sys, TotalAlloc, RandomValue}
	ErrPollerStopped = errors.New("poller: Poller stopped")
)

type Poller interface {
	Run(ctx context.Context) error
}

type MemStatsPoller struct {
	pollInterval time.Duration
	storage      storage.Storage

	mstats runtime.MemStats
}

func NewMemStatsPoller(storage storage.Storage, pollInterval time.Duration) MemStatsPoller {
	return MemStatsPoller{storage: storage, pollInterval: pollInterval}
}

func (sp MemStatsPoller) captureGauge(name storage.MetricName) (storage.Gaugeable, error) {
	switch name {
	// Gauges captured by runtime.ReadMemStats
	case Alloc:
		return storage.Gaugeable(sp.mstats.Alloc), nil
	case BuckHashSys:
		return storage.Gaugeable(sp.mstats.BuckHashSys), nil
	case Frees:
		return storage.Gaugeable(sp.mstats.Frees), nil
	case GCCPUFraction:
		return storage.Gaugeable(sp.mstats.GCCPUFraction), nil
	case GCSys:
		return storage.Gaugeable(sp.mstats.GCSys), nil
	case HeapAlloc:
		return storage.Gaugeable(sp.mstats.HeapAlloc), nil
	case HeapIdle:
		return storage.Gaugeable(sp.mstats.HeapIdle), nil
	case HeapInuse:
		return storage.Gaugeable(sp.mstats.HeapInuse), nil
	case HeapObjects:
		return storage.Gaugeable(sp.mstats.HeapObjects), nil
	case HeapReleased:
		return storage.Gaugeable(sp.mstats.HeapReleased), nil
	case HeapSys:
		return storage.Gaugeable(sp.mstats.HeapSys), nil
	case LastGC:
		return storage.Gaugeable(sp.mstats.LastGC), nil
	case Lookups:
		return storage.Gaugeable(sp.mstats.Lookups), nil
	case MCacheInuse:
		return storage.Gaugeable(sp.mstats.MCacheInuse), nil
	case MCacheSys:
		return storage.Gaugeable(sp.mstats.MCacheSys), nil
	case MSpanInuse:
		return storage.Gaugeable(sp.mstats.MSpanInuse), nil
	case MSpanSys:
		return storage.Gaugeable(sp.mstats.MSpanSys), nil
	case Mallocs:
		return storage.Gaugeable(sp.mstats.Mallocs), nil
	case NextGC:
		return storage.Gaugeable(sp.mstats.NextGC), nil
	case NumForcedGC:
		return storage.Gaugeable(sp.mstats.NumForcedGC), nil
	case NumGC:
		return storage.Gaugeable(sp.mstats.NumGC), nil
	case OtherSys:
		return storage.Gaugeable(sp.mstats.OtherSys), nil
	case PauseTotalNs:
		return storage.Gaugeable(sp.mstats.PauseTotalNs), nil
	case StackInuse:
		return storage.Gaugeable(sp.mstats.StackInuse), nil
	case StackSys:
		return storage.Gaugeable(sp.mstats.StackSys), nil
	case Sys:
		return storage.Gaugeable(sp.mstats.Sys), nil
	case TotalAlloc:
		return storage.Gaugeable(sp.mstats.TotalAlloc), nil
	case RandomValue:
		return storage.Gaugeable(rand.Float64()), nil
	default:
		slog.Error("Unknown metric", "name", name)
		return 0, errors.New("poller: Unknown metric")
	}
}

func (sp MemStatsPoller) captureStats() {
	runtime.ReadMemStats(&sp.mstats)

	for _, name := range gauges {
		value, err := sp.captureGauge(name)

		if err != nil {
			slog.Error("Failed to capture gauge", "name", name, "error", err)
			continue
		}

		sp.storage.SetGauge(name, value)
	}

	sp.storage.UpdateCounter(PollCount, 1)
}

func (sp MemStatsPoller) Run(ctx context.Context) error {
	go func() {
		sp.captureStats()

		ticker := time.NewTicker(sp.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sp.captureStats()
				slog.Info("poller: Metrics updated")
				ticker.Reset(sp.pollInterval)
			}
		}
	}()

	<-ctx.Done()

	return ErrPollerStopped
}
