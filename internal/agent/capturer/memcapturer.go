package capturer

import (
	"log/slog"
	"math/rand"
	"runtime"
	"sync"

	"github.com/nkiryanov/go-metrics/internal/storage"
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
	stor []Stat
}

func NewMemCapturer() *MemCapturer {
	return &MemCapturer{}
}

// Capture mem (mostly) stats
func (c *MemCapturer) Capture() []Stat {
	var ms runtime.MemStats
	var stats = make([]Stat, 0, len(gauges)+len(counters))

	runtime.ReadMemStats(&ms)

	return append(stats,
		[]Stat{
			{Alloc, storage.Gauge(ms.Alloc)},
			{BuckHashSys, storage.Gauge(ms.BuckHashSys)},
			{Frees, storage.Gauge(ms.Frees)},
			{GCCPUFraction, storage.Gauge(ms.GCCPUFraction)},
			{GCSys, storage.Gauge(ms.GCSys)},
			{HeapAlloc, storage.Gauge(ms.HeapAlloc)},
			{HeapIdle, storage.Gauge(ms.HeapIdle)},
			{HeapInuse, storage.Gauge(ms.HeapInuse)},
			{HeapObjects, storage.Gauge(ms.HeapObjects)},
			{HeapReleased, storage.Gauge(ms.HeapReleased)},
			{HeapSys, storage.Gauge(ms.HeapSys)},
			{LastGC, storage.Gauge(ms.LastGC)},
			{Lookups, storage.Gauge(ms.Lookups)},
			{MCacheInuse, storage.Gauge(ms.MCacheInuse)},
			{MCacheSys, storage.Gauge(ms.MCacheSys)},
			{MSpanInuse, storage.Gauge(ms.MSpanInuse)},
			{MSpanSys, storage.Gauge(ms.MSpanSys)},
			{Mallocs, storage.Gauge(ms.Mallocs)},
			{NextGC, storage.Gauge(ms.NextGC)},
			{NumForcedGC, storage.Gauge(ms.NumForcedGC)},
			{NumGC, storage.Gauge(ms.NumGC)},
			{OtherSys, storage.Gauge(ms.OtherSys)},
			{PauseTotalNs, storage.Gauge(ms.PauseTotalNs)},
			{StackInuse, storage.Gauge(ms.StackInuse)},
			{StackSys, storage.Gauge(ms.StackSys)},
			{Sys, storage.Gauge(ms.Sys)},
			{TotalAlloc, storage.Gauge(ms.TotalAlloc)},

			// Capture random gauge
			{RandomValue, storage.Gauge(rand.Float64())},

			// Capture counter
			{PollCount, storage.Counter(1)},
		}...,
	)
}

func (c *MemCapturer) CaptureAndSave() {
	c.mu.Lock()
	c.stor = c.Capture()
	c.mu.Unlock()

	slog.Info("capturer: mem stats saved")
}

func (c *MemCapturer) Last() []Stat {
	c.mu.Lock()
	defer c.mu.Unlock()

	last := make([]Stat, len(c.stor))
	copy(last, c.stor)

	return last
}
