package app

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
	PollInterval time.Duration = 2 * time.Second
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
	ErrAgentStopped = errors.New("agent: Agent stopped")
	gauges          = []storage.MetricName{Alloc, BuckHashSys, Frees, GCCPUFraction, GCSys, HeapAlloc, HeapIdle, HeapInuse, HeapObjects, HeapReleased, HeapSys, LastGC, Lookups, MCacheInuse, MCacheSys, MSpanInuse, MSpanSys, Mallocs, NextGC, NumForcedGC, NumGC, OtherSys, PauseTotalNs, StackInuse, StackSys, Sys, TotalAlloc, RandomValue}
)

type Agent struct {
	storage storage.Storage

	mstats runtime.MemStats
}

func NewAgent(storage storage.Storage) *Agent {
	return &Agent{storage: storage}
}

func (a *Agent) captureGauge(name storage.MetricName) (storage.Gaugeable, error) {
	switch name {
	// Gauges captured by runtime.ReadMemStats
	case Alloc:
		return storage.Gaugeable(a.mstats.Alloc), nil
	case BuckHashSys:
		return storage.Gaugeable(a.mstats.BuckHashSys), nil
	case Frees:
		return storage.Gaugeable(a.mstats.Frees), nil
	case GCCPUFraction:
		return storage.Gaugeable(a.mstats.GCCPUFraction), nil
	case GCSys:
		return storage.Gaugeable(a.mstats.GCSys), nil
	case HeapAlloc:
		return storage.Gaugeable(a.mstats.HeapAlloc), nil
	case HeapIdle:
		return storage.Gaugeable(a.mstats.HeapIdle), nil
	case HeapInuse:
		return storage.Gaugeable(a.mstats.HeapInuse), nil
	case HeapObjects:
		return storage.Gaugeable(a.mstats.HeapObjects), nil
	case HeapReleased:
		return storage.Gaugeable(a.mstats.HeapReleased), nil
	case HeapSys:
		return storage.Gaugeable(a.mstats.HeapSys), nil
	case LastGC:
		return storage.Gaugeable(a.mstats.LastGC), nil
	case Lookups:
		return storage.Gaugeable(a.mstats.Lookups), nil
	case MCacheInuse:
		return storage.Gaugeable(a.mstats.MCacheInuse), nil
	case MCacheSys:
		return storage.Gaugeable(a.mstats.MCacheSys), nil
	case MSpanInuse:
		return storage.Gaugeable(a.mstats.MSpanInuse), nil
	case MSpanSys:
		return storage.Gaugeable(a.mstats.MSpanSys), nil
	case Mallocs:
		return storage.Gaugeable(a.mstats.Mallocs), nil
	case NextGC:
		return storage.Gaugeable(a.mstats.NextGC), nil
	case NumForcedGC:
		return storage.Gaugeable(a.mstats.NumForcedGC), nil
	case NumGC:
		return storage.Gaugeable(a.mstats.NumGC), nil
	case OtherSys:
		return storage.Gaugeable(a.mstats.OtherSys), nil
	case PauseTotalNs:
		return storage.Gaugeable(a.mstats.PauseTotalNs), nil
	case StackInuse:
		return storage.Gaugeable(a.mstats.StackInuse), nil
	case StackSys:
		return storage.Gaugeable(a.mstats.StackSys), nil
	case Sys:
		return storage.Gaugeable(a.mstats.Sys), nil
	case TotalAlloc:
		return storage.Gaugeable(a.mstats.TotalAlloc), nil
	case RandomValue:
		return storage.Gaugeable(rand.Float64()), nil
	default:
		slog.Error("Unknown metric", "name", name)
		return 0, errors.New("agent: Unknown metric")
	}
}

func (a *Agent) captureStats() {
	runtime.ReadMemStats(&a.mstats)

	for _, name := range gauges {
		value, err := a.captureGauge(name)

		if err != nil {
			slog.Error("Failed to capture gauge", "name", name, "error", err)
			continue
		}

		a.storage.SetGauge(name, value)
	}

	a.storage.UpdateCounter(PollCount, 1)
}

func (a *Agent) Poll(ctx context.Context) error {
	go func() {
		a.captureStats()

		ticker := time.NewTicker(PollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.captureStats()
				slog.Info("Metrics updated")
				ticker.Reset(PollInterval)
			}
		}
	}()

	<-ctx.Done()

	return ErrAgentStopped
}
