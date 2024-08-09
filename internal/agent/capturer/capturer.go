package capturer

import (
	"github.com/nkiryanov/go-metrics/internal/storage"
)

type Stat struct {
	Name  string
	Value storage.Storable
}

type Capturer interface {
	Capture() []Stat
}
