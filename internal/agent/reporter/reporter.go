package reporter

import (
	"fmt"
)

type Metric struct {
	Name  string
	Type  string
	Value fmt.Stringer
}

type Reporter interface {
	ReportOnce(m *Metric) error
	ReportBatch(ms []*Metric) []error
}
