package handlers

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/nkiryanov/go-metrics/internal/handlers/templates"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

type MetricsAPIHandler interface {
	RegisterRoutes(chi.Router)
}

type MetricsAPI struct {
	storage storage.Storage

	listTpl *template.Template
}

func NewMetricsAPI(storage storage.Storage) MetricsAPI {
	return MetricsAPI{
		storage: storage,
		listTpl: templates.MetricList,
	}
}

func (api MetricsAPI) RegisterRoutes(r chi.Router) {
	r.With(middleware.SetHeader("Content-Type", "text/html")).Get("/", api.listMetrics)
	r.With(middleware.SetHeader("Content-Type", "text/plain")).Get("/value/{mType}/{mName}", api.getMetric)

	r.Route("/update", func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "text/plain"))
		r.Post("/counter/{mName}/{mValue}", api.updateCounter)
		r.Post("/gauge/{mName}/{mValue}", api.updateGauge)
		r.Post("/{mType}/{mName}/{mValue}", func(w http.ResponseWriter, r *http.Request) { http.Error(w, "Bad Request", http.StatusBadRequest) })
	})
}

func writeValue(w http.ResponseWriter, value string) {
	if _, err := w.Write([]byte(value)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (api MetricsAPI) updateCounter(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("mName")

	countable, err := strconv.ParseInt(r.PathValue("mValue"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	storedValue := api.storage.UpdateCounter(storage.MetricName(name), storage.Countable(countable))
	slog.Info("Counter updated", "name", name, "value", storedValue)

	writeValue(w, storedValue.String())
}

func (api MetricsAPI) updateGauge(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("mName")

	gauge, err := strconv.ParseFloat(r.PathValue("mValue"), 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	storedValue := api.storage.SetGauge(storage.MetricName(name), storage.Gaugeable(gauge))
	slog.Info("Gauge updated", "name", name, "value", storedValue)

	writeValue(w, storedValue.String())
}

func (api MetricsAPI) getMetric(w http.ResponseWriter, r *http.Request) {
	mType := r.PathValue("mType")
	name := r.PathValue("mName")

	var value string
	var ok bool

	switch mType {
	case string(storage.CounterTypeName):
		var counter storage.Countable
		counter, ok = api.storage.GetCounter(storage.MetricName(name))
		if ok {
			value = counter.String()
		}
	case string(storage.GaugeTypeName):
		var gauge storage.Gaugeable
		gauge, ok = api.storage.GetGauge(storage.MetricName(name))
		if ok {
			value = gauge.String()
		}
	default:
		ok = false
	}

	if !ok {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	writeValue(w, value)
}

func (api MetricsAPI) listMetrics(w http.ResponseWriter, _ *http.Request) {
	metrics := api.storage.ListMetrics()

	if err := api.listTpl.Execute(w, metrics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
