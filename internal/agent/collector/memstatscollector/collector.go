package memstatscollector

import (
	"context"
	"math/rand"
	"runtime"

	"github.com/nkiryanov/go-metrics/internal/agent/collector/memstorage"
	"github.com/nkiryanov/go-metrics/internal/models"
)

const (
	// Gauges captured by runtime.ReadMemStats
	Alloc         = "Alloc"
	BuckHashSys   = "BuckHashSys"
	Frees         = "Frees"
	GCCPUFraction = "GCCPUFraction"
	GCSys         = "GCSys"
	HeapAlloc     = "HeapAlloc"
	HeapIdle      = "HeapIdle"
	HeapInuse     = "HeapInuse"
	HeapObjects   = "HeapObjects"
	HeapReleased  = "HeapReleased"
	HeapSys       = "HeapSys"
	LastGC        = "LastGC"
	Lookups       = "Lookups"
	MCacheInuse   = "MCacheInuse"
	MCacheSys     = "MCacheSys"
	MSpanInuse    = "MSpanInuse"
	MSpanSys      = "MSpanSys"
	Mallocs       = "Mallocs"
	NextGC        = "NextGC"
	NumForcedGC   = "NumForcedGC"
	NumGC         = "NumGC"
	OtherSys      = "OtherSys"
	PauseTotalNs  = "PauseTotalNs"
	StackInuse    = "StackInuse"
	StackSys      = "StackSys"
	Sys           = "Sys"
	TotalAlloc    = "TotalAlloc"
	RandomValue   = "RandomValue"

	// Counters computed by the agent
	PollCount = "PollCount"
)

// Collect metrics from 'runtime.ReadMemStats' module (and some other that we close eyes on)
type MemStatsCollector struct {
	storage *memstorage.MemStorage // Simplest possible storage
}

func New() *MemStatsCollector {
	return &MemStatsCollector{
		storage: memstorage.New(),
	}
}

func (c *MemStatsCollector) List() []models.Metric {
	return c.storage.List()
}

func (c *MemStatsCollector) Collect(_ context.Context) error {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	c.storage.Set(
		[]models.Metric{
			{Name: Alloc, Type: models.GaugeTypeName, Value: float64(ms.Alloc)},
			{Name: BuckHashSys, Type: models.GaugeTypeName, Value: float64(ms.BuckHashSys)},
			{Name: Frees, Type: models.GaugeTypeName, Value: float64(ms.Frees)},
			{Name: GCCPUFraction, Type: models.GaugeTypeName, Value: float64(ms.GCCPUFraction)},
			{Name: GCSys, Type: models.GaugeTypeName, Value: float64(ms.GCSys)},
			{Name: HeapAlloc, Type: models.GaugeTypeName, Value: float64(ms.HeapAlloc)},
			{Name: HeapIdle, Type: models.GaugeTypeName, Value: float64(ms.HeapIdle)},
			{Name: HeapInuse, Type: models.GaugeTypeName, Value: float64(ms.HeapInuse)},
			{Name: HeapObjects, Type: models.GaugeTypeName, Value: float64(ms.HeapObjects)},
			{Name: HeapReleased, Type: models.GaugeTypeName, Value: float64(ms.HeapReleased)},
			{Name: HeapSys, Type: models.GaugeTypeName, Value: float64(ms.HeapSys)},
			{Name: LastGC, Type: models.GaugeTypeName, Value: float64(ms.LastGC)},
			{Name: Lookups, Type: models.GaugeTypeName, Value: float64(ms.Lookups)},
			{Name: MCacheInuse, Type: models.GaugeTypeName, Value: float64(ms.MCacheInuse)},
			{Name: MCacheSys, Type: models.GaugeTypeName, Value: float64(ms.MCacheSys)},
			{Name: MSpanInuse, Type: models.GaugeTypeName, Value: float64(ms.MSpanInuse)},
			{Name: MSpanSys, Type: models.GaugeTypeName, Value: float64(ms.MSpanSys)},
			{Name: Mallocs, Type: models.GaugeTypeName, Value: float64(ms.Mallocs)},
			{Name: NextGC, Type: models.GaugeTypeName, Value: float64(ms.NextGC)},
			{Name: NumForcedGC, Type: models.GaugeTypeName, Value: float64(ms.NumForcedGC)},
			{Name: NumGC, Type: models.GaugeTypeName, Value: float64(ms.NumGC)},
			{Name: OtherSys, Type: models.GaugeTypeName, Value: float64(ms.OtherSys)},
			{Name: PauseTotalNs, Type: models.GaugeTypeName, Value: float64(ms.PauseTotalNs)},
			{Name: StackInuse, Type: models.GaugeTypeName, Value: float64(ms.StackInuse)},
			{Name: StackSys, Type: models.GaugeTypeName, Value: float64(ms.StackSys)},
			{Name: Sys, Type: models.GaugeTypeName, Value: float64(ms.Sys)},
			{Name: TotalAlloc, Type: models.GaugeTypeName, Value: float64(ms.TotalAlloc)},

			// Capture random gauge
			{Name: RandomValue, Type: models.GaugeTypeName, Value: rand.Float64()},

			// Capture counter
			{Name: PollCount, Type: models.CounterTypeName, Delta: 1},
		}...,
	)

	return nil
}
