package reporter

import (
	"github.com/nkiryanov/go-metrics/internal/models"
)

type Reporter interface {
	// Should report metric
	// Return error if any error occurs
	ReportOnce(m *models.Metric) error

	// Should iterate ms and report all of them
	ReportBatch(ms []models.Metric) error
}
