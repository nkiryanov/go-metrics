package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

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

	router.Use(LoggerMiddleware, GzipMiddleware)

	// Root level
	router.Get("/", listMetrics(stor, templates.MetricList))

	// Routers for /value
	router.Route("/value", func(router chi.Router) {
		router.Get("/{mType}/{mID}", getMetricPlain(stor))
		router.Post("/", getMetricJSON(stor))
	})

	// Routers for /update
	router.Route("/update", func(router chi.Router) {
		router.Post("/", updateMetricJSON(stor))
		router.Post("/{mType}/{mID}/{mValue}", updateMetricPlain(stor))
	})

	return router
}
