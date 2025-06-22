package gopsutilcollector

import (
	"context"
	"fmt"

	"github.com/nkiryanov/go-metrics/internal/agent/collector/memstorage"
	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"golang.org/x/sync/errgroup"
)

const (
	TotalMemory    = "TotalMemory"
	FreeMemory     = "FreeMemory"
	CPUutilization = "CPUutilization"
)

// Collect metrics from 'runtime.ReadMemStats' module (and some other that we close eyes on)
type GopsutilCollector struct {
	storage *memstorage.MemStorage // Simplest possible storage
}

func New() *GopsutilCollector {
	return &GopsutilCollector{
		storage: memstorage.New(),
	}
}

func (c *GopsutilCollector) List() []models.Metric {
	return c.storage.List()
}

func (c *GopsutilCollector) Collect(_ context.Context) error {
	var eg errgroup.Group

	// Collect memory metrics
	memMetrics := make([]models.Metric, 0, 2)
	eg.Go(func() error {
		vmStats, err := mem.VirtualMemory()
		if err != nil {
			return err
		}

		memMetrics = append(
			memMetrics,
			models.Metric{Name: TotalMemory, Type: "gauge", Value: float64(vmStats.Total)},
			models.Metric{Name: FreeMemory, Type: "gauge", Value: float64(vmStats.Free)},
		)

		return nil
	})

	// Collect cpu utilization
	cpuMetrics := make([]models.Metric, 0)
	eg.Go(func() error {
		cpuStats, err := cpu.Percent(0, true) // per cpu utilization
		if err != nil {
			return err
		}

		for cpuNum, utilization := range cpuStats {
			cpuMetrics = append(
				cpuMetrics,
				models.Metric{
					Name:  fmt.Sprintf("%s%d", CPUutilization, cpuNum+1),
					Type:  "gauge",
					Value: utilization,
				},
			)
		}

		return nil
	})

	err := eg.Wait()
	c.storage.Set(append(memMetrics, cpuMetrics...)...)
	if err != nil {
		return err
	}

	return nil
}
