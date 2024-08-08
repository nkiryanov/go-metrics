package handlers

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

func updateMetric(s storage.Storage, parser storage.StorableParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, URLMetricType)
		mName := chi.URLParam(r, URLMetricName)
		mValue := chi.URLParam(r, URLMetricValue)

		var err error
		var storable storage.Storable
		storable, err = parser.Parse(mType, mValue)
		if err != nil {
			slog.Warn("bad request for update metric", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		storable, err = s.UpdateMetric(mName, storable)
		if err != nil {
			slog.Warn("can't update metric", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		slog.Info("Metric updated", "type", mType, "name", mName, "value", storable)
		writeOrInternalError(w, storable.String())
	}
}

func getMetric(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := r.PathValue("mType")
		mName := r.PathValue("mName")

		storable, ok, err := s.GetMetric(mType, mName)
		if err != nil {
			slog.Error("storage error occurred when metric requested", "error", err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if !ok {
			slog.Info("metic requested, but not found", "type", mType, "name", mName)
			http.Error(w, fmt.Sprintf("metric not found. type: %s, name: %s", mType, mName), http.StatusNotFound)
			return
		}

		writeOrInternalError(w, storable.String())
	}
}

func listMetrics(s storage.Storage, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type metric struct {
			Type  string
			Name  string
			Value string
		}

		metrics := make([]metric, 0, s.Len())

		s.Iterate(func(mt string, mn string, value storage.Storable) {
			metrics = append(metrics, metric{mt, mn, value.String()})
		})

		if err := tpl.Execute(w, metrics); err != nil {
			slog.Error("list metric templated generation failed", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func writeOrInternalError(w http.ResponseWriter, value string) {
	_, err := w.Write([]byte(value))
	if err != nil {
		slog.Error("write response error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
