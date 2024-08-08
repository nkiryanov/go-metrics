package capturer

import (
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
	gauges           = []string{Alloc, BuckHashSys, Frees, GCCPUFraction, GCSys, HeapAlloc, HeapIdle, HeapInuse, HeapObjects, HeapReleased, HeapSys, LastGC, Lookups, MCacheInuse, MCacheSys, MSpanInuse, MSpanSys, Mallocs, NextGC, NumForcedGC, NumGC, OtherSys, PauseTotalNs, StackInuse, StackSys, Sys, TotalAlloc, RandomValue}
	counters		 = []string{PollCount}
)

type MemCapturer struct {
	mu sync.Mutex
	mstats runtime.MemStats
}

func NewMemCapturer() *MemCapturer {
	return &MemCapturer{}
}

// Capture mem (mostly) stats
func (c *MemCapturer) Capture() []Stat {
	var stats = make([]Stat, 0, len(gauges) + len(counters))

	c.mu.Lock()
	defer c.mu.Unlock()

	runtime.ReadMemStats(&c.mstats)

	return append(stats, 
		[]Stat{
			{Alloc, storage.Gauge(c.mstats.Alloc)},
			{BuckHashSys, storage.Gauge(c.mstats.BuckHashSys)},
			{Frees, storage.Gauge(c.mstats.Frees)},
			{GCCPUFraction, storage.Gauge(c.mstats.GCCPUFraction)},
			{GCSys, storage.Gauge(c.mstats.GCSys)},
			{HeapAlloc, storage.Gauge(c.mstats.HeapAlloc)},
			{HeapIdle, storage.Gauge(c.mstats.HeapIdle)},
			{HeapInuse, storage.Gauge(c.mstats.HeapInuse)},
			{HeapObjects, storage.Gauge(c.mstats.HeapObjects)},
			{HeapReleased, storage.Gauge(c.mstats.HeapReleased)},
			{HeapSys, storage.Gauge(c.mstats.HeapSys)},
			{LastGC, storage.Gauge(c.mstats.LastGC)},
			{Lookups, storage.Gauge(c.mstats.Lookups)},
			{MCacheInuse, storage.Gauge(c.mstats.MCacheInuse)},
			{MCacheSys, storage.Gauge(c.mstats.MCacheSys)},
			{MSpanInuse, storage.Gauge(c.mstats.MSpanInuse)},
			{MSpanSys, storage.Gauge(c.mstats.MSpanSys)},
			{Mallocs, storage.Gauge(c.mstats.Mallocs)},
			{NextGC, storage.Gauge(c.mstats.NextGC)},
			{NumForcedGC, storage.Gauge(c.mstats.NumForcedGC)},
			{NumGC, storage.Gauge(c.mstats.NumGC)},
			{OtherSys, storage.Gauge(c.mstats.OtherSys)},
			{PauseTotalNs, storage.Gauge(c.mstats.PauseTotalNs)},
			{StackInuse, storage.Gauge(c.mstats.StackInuse)},
			{StackSys, storage.Gauge(c.mstats.StackSys)},
			{Sys, storage.Gauge(c.mstats.Sys)},
			{TotalAlloc, storage.Gauge(c.mstats.TotalAlloc)},

			// Capture random gauge
			{RandomValue, storage.Gauge(rand.Float64())},

			// Capture counter
			{PollCount, storage.Counter(1)},
		}...
	)
}
