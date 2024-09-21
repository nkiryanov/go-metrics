package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/models"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

func updateMetricPlain(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := chi.URLParam(r, URLMetricType)
		mID := chi.URLParam(r, URLMetricID)
		mValue := chi.URLParam(r, URLMetricValue)

		var err error
		var msg string
		metric := models.Metric{ID: mID, MType: mType}

		switch mType {
		case models.CounterTypeName:
			if metric.Delta, err = strconv.ParseInt(mValue, 10, 64); err != nil {
				msg = "bad value to update counter"
			}
		case models.GaugeTypeName:
			if metric.Value, err = strconv.ParseFloat(mValue, 64); err != nil {
				msg = "bad value to update gauge"
			}
		default:
			msg = "bad metric type"
			err = errors.New(msg)
		}

		if err != nil {
			logger.Slog.Warnw(msg, "error", err)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		if metric, err = s.UpdateMetric(&metric); err != nil {
			logger.Slog.Warnw("can't update metric", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Slog.Infow("Metric updated with", "id", mID, "type", mType, "value", mValue)
		w.Header().Set("Content-Type", "text/plain")
		writeOrInternalError(w, []byte(metric.String()))
	}
}

func updateMetricJSON(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var metric models.Metric

		if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
			logger.Slog.Warnw("request not expected format", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		metric, err := s.UpdateMetric(&metric)
		if err != nil {
			logger.Slog.Warnw("metric could not updated", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Slog.Infow("Metric updated", "id", metric.ID, "type", metric.MType, "value", metric.String())

		resp, err := json.Marshal(models.NewMetricJSON(&metric))
		if err != nil {
			logger.Slog.Error("error while deserializing metric json", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		writeOrInternalError(w, resp)
	}
}

func getMetricPlain(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mType := r.PathValue("mType")
		mID := r.PathValue("mID")

		metric, ok, err := s.GetMetric(mID, mType)
		if err != nil {
			logger.Slog.Error("storage error occurred when metric requested", "error", err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if !ok {
			logger.Slog.Info("metic requested, but not found", "type", mType, "id", mID)
			http.Error(w, fmt.Sprintf("metric not found. type: %s, id: %s", mType, mID), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		writeOrInternalError(w, []byte(metric.String()))
	}
}

func getMetricJSON(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &struct {
			ID    string `json:"id"`
			MType string `json:"type"`
		}{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Slog.Warnw("error while decoding request", "error", err.Error())
			http.Error(w, "request not in expected format", http.StatusBadRequest)
			return
		}

		metric, ok, err := s.GetMetric(req.ID, req.MType)
		if err != nil {
			logger.Slog.Error("storage error occurred when metric requested", "error", err.Error())
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if !ok {
			logger.Slog.Info("metic requested, but not found", "type", req.MType, "id", req.ID)
			http.Error(w, fmt.Sprintf("metric not found. type: %s, id: %s", req.MType, req.ID), http.StatusNotFound)
			return
		}

		resp, err := json.Marshal(models.NewMetricJSON(&metric))
		if err != nil {
			logger.Slog.Error("error while deserializing metric json", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		writeOrInternalError(w, resp)
	}
}

func listMetrics(s storage.Storage, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type metric struct {
			ID    string
			Type  string
			Value string
		}

		metrics := make([]metric, 0, s.Count())

		_ = s.Iterate(func(m models.Metric) error {
			metrics = append(metrics, metric{ID: m.ID, Type: m.MType, Value: m.String()})
			return nil
		})

		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].ID < metrics[j].ID
		})

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		if err := tpl.Execute(w, metrics); err != nil {
			logger.Slog.Error("list metric templated generation failed", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func writeOrInternalError(w http.ResponseWriter, value []byte) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(value)

	if err != nil {
		logger.Slog.Error("write response error", "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
