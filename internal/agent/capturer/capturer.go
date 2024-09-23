package capturer

import (
	"github.com/nkiryanov/go-metrics/internal/models"
)

// Capturer do what the names mean: it capture stats
type Capturer interface {
	// Capture the statistics and returns them as an array of Stat objects.
	// This method should be thread-safe.
	Capture() []models.Metric

	// Capture and save stats.
	CaptureAndSave()

	// Return last saved stats
	Last() []models.Metric
}
