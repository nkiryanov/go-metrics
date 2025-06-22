package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/nkiryanov/go-metrics/internal/logger"
	"github.com/nkiryanov/go-metrics/internal/server/handlers/templates"
	"github.com/nkiryanov/go-metrics/internal/server/storage"
)

const (
	URLMetricID    string = "mID"
	URLMetricType  string = "mType"
	URLMetricValue string = "mValue"
)

func NewMetricRouter(stor storage.Storage, lgr logger.Logger, hmacSecretKey string) http.Handler {
	router := chi.NewRouter()

	router.Use(
		LoggerMiddleware(lgr),
		HmacSHA256Middleware(lgr, hmacSecretKey),
		GzipMiddleware,
	)

	// Root level
	router.Get("/", listMetrics(stor, lgr, templates.MetricList))

	// Routers for /value
	router.Route("/value", func(router chi.Router) {
		router.Get("/{mType}/{mID}", getMetricPlain(stor, lgr))
		router.Post("/", getMetricJSON(stor, lgr))
	})

	// Routers for /update
	router.Route("/update", func(router chi.Router) {
		router.Post("/", updateMetricJSON(stor, lgr))
		router.Post("/{mType}/{mID}/{mValue}", updateMetricPlain(stor, lgr))
	})

	// Router for /updates
	router.Route("/updates", func(router chi.Router) {
		router.Post("/", updateMetricBulkJSON(stor, lgr))
	})

	// Routers for /ping
	router.Get("/ping", ping(stor, lgr))

	return router
}
