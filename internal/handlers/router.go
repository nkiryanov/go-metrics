package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/nkiryanov/go-metrics/internal/handlers/templates"
	"github.com/nkiryanov/go-metrics/internal/storage"
)

const (
	URLMetricID    string = "mID"
	URLMetricType  string = "mType"
	URLMetricValue string = "mValue"
)

func NewMetricRouter(stor storage.Storage) http.Handler {
	router := chi.NewRouter()

	router.Use(LoggerMiddleware)

	// Always allowed despite the content type
	router.With(middleware.SetHeader("Content-Type", "text/html")).Get("/", listMetrics(stor, templates.MetricList))

	// Routers for /value
	router.Route("/value", func(router chi.Router) {
		router.With(middleware.SetHeader("Content-Type", "text/plain")).Get("/{mType}/{mID}", getMetricPlain(stor))
		router.With(middleware.SetHeader("Content-Type", "application/json")).Post("/", getMetricJSON(stor))
	})

	// Routers for /update
	router.Route("/update", func(router chi.Router) {
		router.With(middleware.SetHeader("Content-Type", "text/plain")).Post("/{mType}/{mID}/{mValue}", updateMetricPlain(stor))
	})

	return router
}
