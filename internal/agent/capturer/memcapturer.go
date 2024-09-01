package capturer

import (
	"math/rand"
	"runtime"
	"sync"

	"github.com/nkiryanov/go-metrics/internal/logger"
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

var (
	gauges   = []string{Alloc, BuckHashSys, Frees, GCCPUFraction, GCSys, HeapAlloc, HeapIdle, HeapInuse, HeapObjects, HeapReleased, HeapSys, LastGC, Lookups, MCacheInuse, MCacheSys, MSpanInuse, MSpanSys, Mallocs, NextGC, NumForcedGC, NumGC, OtherSys, PauseTotalNs, StackInuse, StackSys, Sys, TotalAlloc, RandomValue}
	counters = []string{PollCount}
)

type MemCapturer struct {
	mu   sync.Mutex
	stor []models.Metric
}

func NewMemCapturer() *MemCapturer {
	return &MemCapturer{}
}

// Capture mem (mostly) stats
func (c *MemCapturer) Capture() []models.Metric {
	var ms runtime.MemStats
	var stats = make([]models.Metric, 0, len(gauges)+len(counters))

	runtime.ReadMemStats(&ms)

	return append(stats,
		[]models.Metric{
			{ID: Alloc, MType: models.GaugeTypeName, Value: float64(ms.Alloc)},
			{ID: BuckHashSys, MType: models.GaugeTypeName, Value: float64(ms.BuckHashSys)},
			{ID: Frees, MType: models.GaugeTypeName, Value: float64(ms.Frees)},
			{ID: GCCPUFraction, MType: models.GaugeTypeName, Value: float64(ms.GCCPUFraction)},
			{ID: GCSys, MType: models.GaugeTypeName, Value: float64(ms.GCSys)},
			{ID: HeapAlloc, MType: models.GaugeTypeName, Value: float64(ms.HeapAlloc)},
			{ID: HeapIdle, MType: models.GaugeTypeName, Value: float64(ms.HeapIdle)},
			{ID: HeapInuse, MType: models.GaugeTypeName, Value: float64(ms.HeapInuse)},
			{ID: HeapObjects, MType: models.GaugeTypeName, Value: float64(ms.HeapObjects)},
			{ID: HeapReleased, MType: models.GaugeTypeName, Value: float64(ms.HeapReleased)},
			{ID: HeapSys, MType: models.GaugeTypeName, Value: float64(ms.HeapSys)},
			{ID: LastGC, MType: models.GaugeTypeName, Value: float64(ms.LastGC)},
			{ID: Lookups, MType: models.GaugeTypeName, Value: float64(ms.Lookups)},
			{ID: MCacheInuse, MType: models.GaugeTypeName, Value: float64(ms.MCacheInuse)},
			{ID: MCacheSys, MType: models.GaugeTypeName, Value: float64(ms.MCacheSys)},
			{ID: MSpanInuse, MType: models.GaugeTypeName, Value: float64(ms.MSpanInuse)},
			{ID: MSpanSys, MType: models.GaugeTypeName, Value: float64(ms.MSpanSys)},
			{ID: Mallocs, MType: models.GaugeTypeName, Value: float64(ms.Mallocs)},
			{ID: NextGC, MType: models.GaugeTypeName, Value: float64(ms.NextGC)},
			{ID: NumForcedGC, MType: models.GaugeTypeName, Value: float64(ms.NumForcedGC)},
			{ID: NumGC, MType: models.GaugeTypeName, Value: float64(ms.NumGC)},
			{ID: OtherSys, MType: models.GaugeTypeName, Value: float64(ms.OtherSys)},
			{ID: PauseTotalNs, MType: models.GaugeTypeName, Value: float64(ms.PauseTotalNs)},
			{ID: StackInuse, MType: models.GaugeTypeName, Value: float64(ms.StackInuse)},
			{ID: StackSys, MType: models.GaugeTypeName, Value: float64(ms.StackSys)},
			{ID: Sys, MType: models.GaugeTypeName, Value: float64(ms.Sys)},
			{ID: TotalAlloc, MType: models.GaugeTypeName, Value: float64(ms.TotalAlloc)},

			// Capture random gauge
			{ID: RandomValue, MType: models.GaugeTypeName, Value: rand.Float64()},

			// Capture counter
			{ID: PollCount, MType: models.CounterTypeName, Value: 1},
		}...,
	)
}

func (c *MemCapturer) CaptureAndSave() {
	c.mu.Lock()
	c.stor = c.Capture()
	c.mu.Unlock()

	logger.Slog.Info("capturer: mem stats saved")
}

func (c *MemCapturer) Last() []models.Metric {
	c.mu.Lock()
	defer c.mu.Unlock()

	last := make([]models.Metric, len(c.stor))
	copy(last, c.stor)

	return last
}
