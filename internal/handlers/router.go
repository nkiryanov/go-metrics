package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/nkiryanov/go-metrics/internal/handlers/templates"
	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

const (
	URLMetricType  string = "mType"
	URLMetricName  string = "mName"
	URLMetricValue string = "mValue"
)

func NewMetricRouter(stor storage.Storage, parser storage.StorableParser) http.Handler {
	router := chi.NewRouter()

	router.Use(logger.RequestLogger)

	router.With(middleware.SetHeader("Content-Type", "text/html")).Get("/", listMetrics(stor, templates.MetricList))
	router.With(middleware.SetHeader("Content-Type", "text/plain")).Get("/value/{mType}/{mName}", getMetric(stor))
	router.Route("/update", func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "text/plain"))
		r.Post("/{mType}/{mName}/{mValue}", updateMetric(stor, parser))
	})

	return router
}
