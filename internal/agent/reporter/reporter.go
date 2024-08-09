package reporter

import (
	"fmt"
)

//go:generate moq -out mocks/reporter.go -pkg mocks -skip-ensure -fmt goimports . Reporter

type Metric struct {
	Name  string
	Type  string
	Value fmt.Stringer
}

type Reporter interface {
	// Should report metric
	// Return error if any error occurs
	ReportOnce(m *Metric) error

	// Should iterate ms and report all of them
	ReportBatch(ms []*Metric) error
}
