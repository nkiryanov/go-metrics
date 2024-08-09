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
	ReportOnce(m *Metric) error
	ReportBatch(ms []*Metric) []error
}
