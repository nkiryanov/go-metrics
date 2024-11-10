package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

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
		metric := models.Metric{Name: mID, Type: mType}

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

		if metric, err = s.UpdateMetric(r.Context(), &metric); err != nil {
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

		metric, err := s.UpdateMetric(r.Context(), &metric)
		if err != nil {
			logger.Slog.Warnw("metric could not updated", "error", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		logger.Slog.Infow("Metric updated", "name", metric.Name, "type", metric.Type, "value", metric.String())

		resp, err := json.Marshal(metric)
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
		mName := r.PathValue("mID")

		metric, err := s.GetMetric(r.Context(), mType, mName)
		if err != nil {
			logger.Slog.Info("metic requested, but not found", "type", mType, "name", mName)
			http.Error(w, fmt.Sprintf("metric not found. type: %s, id: %s", mType, mName), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		writeOrInternalError(w, []byte(metric.String()))
	}
}

func getMetricJSON(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &struct {
			Type string `json:"type"`
			Name string `json:"id"`
		}{}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Slog.Warnw("error while decoding request", "error", err.Error())
			http.Error(w, "request not in expected format", http.StatusBadRequest)
			return
		}

		metric, err := s.GetMetric(r.Context(), req.Type, req.Name)
		if err != nil {
			logger.Slog.Infow("metic requested, but not found", "type", req.Type, "id", req.Name)
			http.Error(w, fmt.Sprintf("metric not found. type: %s, id: %s", req.Type, req.Name), http.StatusNotFound)
			return
		}

		resp, err := json.Marshal(metric)
		if err != nil {
			logger.Slog.Errorw("error while deserializing metric json", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		writeOrInternalError(w, resp)
	}
}

func listMetrics(s storage.Storage, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type templateEntry struct {
			ID    string
			Type  string
			Value string
		}

		metrics, err := s.ListMetric(r.Context())
		if err != nil {
			logger.Slog.Infow("list metric failed", "error", err.Error())
		}

		entries := make([]templateEntry, 0, len(metrics))
		for _, m := range metrics {
			entries = append(entries, templateEntry{ID: m.Name, Type: m.Type, Value: m.String()})
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		if err := tpl.Execute(w, entries); err != nil {
			logger.Slog.Errorw("list metric templated generation failed", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func ping(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		if err := s.Ping(ctx); err != nil {
			logger.Slog.Error("db connection failed", "error", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		writeOrInternalError(w, []byte("OK"))
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
