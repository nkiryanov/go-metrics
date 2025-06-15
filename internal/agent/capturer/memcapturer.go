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
	gauges = []string{
		Alloc,
		BuckHashSys,
		Frees,
		GCCPUFraction,
		GCSys,
		HeapAlloc,
		HeapIdle,
		HeapInuse,
		HeapObjects,
		HeapReleased,
		HeapSys,
		LastGC,
		Lookups,
		MCacheInuse,
		MCacheSys,
		MSpanInuse,
		MSpanSys,
		Mallocs,
		NextGC,
		NumForcedGC,
		NumGC,
		OtherSys,
		PauseTotalNs,
		StackInuse,
		StackSys,
		Sys,
		TotalAlloc,
		RandomValue,
	}
	counters = []string{
		PollCount,
	}
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
}

func (c *MemCapturer) CaptureAndSave() {
	c.mu.Lock()
	c.stor = c.Capture()
	c.mu.Unlock()

	logger.Slog.Info("capturer: mem stats saved")
}

func (c *MemCapturer) ListLast() []models.Metric {
	c.mu.Lock()
	defer c.mu.Unlock()

	last := make([]models.Metric, len(c.stor))
	copy(last, c.stor)

	return last
}
