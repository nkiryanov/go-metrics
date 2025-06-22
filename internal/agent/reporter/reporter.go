package reporter

import (
	"github.com/nkiryanov/go-metrics/internal/models"
)

//go:generate moq -out mocks/reporter.go -pkg mocks -skip-ensure -fmt goimports . Reporter

type Reporter interface {
	// Should report metric
	// Return error if any error occurs
	ReportOnce(metric models.Metric) error

	// Should iterate ms and report all of them
	ReportBatch(metrics []models.Metric) error
}
