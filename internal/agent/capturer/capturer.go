package capturer

import (
	"github.com/nkiryanov/go-metrics/internal/storage"
)

type Stat struct {
	Name  string
	Value storage.Storable
}

// Capturer do what the names mean: it capture stats
type Capturer interface {
	// Capture the statistics and returns them as an array of Stat objects.
	// This method should be thread-safe.
	Capture() []Stat

	// Capture and save stats.
	CaptureAndSave()

	// Return last saved stats
	Last() []Stat
}
