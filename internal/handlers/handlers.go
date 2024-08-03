package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/nkiryanov/go-metrics/internal/storage"
)

type MetricsAPIHandler interface {
	UpdateCounter(w http.ResponseWriter, r *http.Request)
	UpdateGauge(w http.ResponseWriter, r *http.Request)
}

type MetricsAPI struct {
	storage storage.Storage
}

func NewMetricsAPI(storage storage.Storage) MetricsAPI {
	return MetricsAPI{storage: storage}
}

func (api MetricsAPI) UpdateCounter(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("mName")

	countable, err := strconv.ParseInt(r.PathValue("mValue"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stored := api.storage.UpdateCounter(storage.MetricName(name), storage.Countable(countable))
	slog.Info("Counter updated", "name", name, "value", stored)
}

func (api MetricsAPI) UpdateGauge(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("mName")

	gauge, err := strconv.ParseFloat(r.PathValue("mValue"), 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	stored := api.storage.SetGauge(storage.MetricName(name), storage.Gaugeable(gauge))
	slog.Info("Gauge updated", "name", name, "value", stored)
}
