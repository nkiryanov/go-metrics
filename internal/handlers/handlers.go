package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/nkiryanov/go-metrics/internal/storage"
)

type MetricsAPI struct {
	storage *storage.MemStorage
}

func NewMetricsAPI(storage *storage.MemStorage) MetricsAPI {
	return MetricsAPI{storage: storage}
}

func (api *MetricsAPI) UpdateCounter(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("metricName")

	countable, err := strconv.ParseInt(r.PathValue("value"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stored := api.storage.UpdateCounter(storage.MetricName(name), storage.Countable(countable))
	slog.Info("Counter updated", "name", name, "value", stored)
}

func (api *MetricsAPI) UpdateGauge(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("metricName")

	gauge, err := strconv.ParseFloat(r.PathValue("value"), 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stored := api.storage.SetGauge(storage.MetricName(name), storage.Gaugeable(gauge))
	slog.Info("Gauge updated", "name", name, "value", stored)
}
