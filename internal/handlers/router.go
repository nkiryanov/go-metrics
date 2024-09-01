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

	router.With(middleware.SetHeader("Content-Type", "text/html")).Get("/", listMetrics(stor, templates.MetricList))
	router.With(middleware.SetHeader("Content-Type", "text/plain")).Get("/value/{mType}/{mID}", getMetricPlain(stor))
	router.Route("/update", func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "text/plain"))
		r.Post("/{mType}/{mID}/{mValue}", updateMetricPlain(stor))
	})

	return router
}
