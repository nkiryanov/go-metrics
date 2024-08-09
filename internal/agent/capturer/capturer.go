package capturer

import (
	"github.com/nkiryanov/go-metrics/internal/storage"
)


//go:generate moq -out mocks/capturer.go -pkg mocks -skip-ensure -fmt goimports . Capturer

type Stat struct {
	Name  string
	Value storage.Storable
}

// Capturer do what the names mean: it capture stats
type Capturer interface {
	// Should only capture stats. Should be thread safe
	Capture() []Stat

	// Should capture (same method as Capture()) and save []Stat in 's storage.Storage'
	// Return error if storage not support any of stat
	CaptureWithSave(s storage.Storage) error
}
