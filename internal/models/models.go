package models

import (
	"strconv"
)

const (
	CounterTypeName = "counter"
	GaugeTypeName   = "gauge"
)

type Metric struct {
	ID    string  `json:"id"`
	MType string  `json:"type"`
	Delta int64   `json:"delta"`
	Value float64 `json:"value"`
}

func (m *Metric) String() string {
	switch m.MType {
	case CounterTypeName:
		return strconv.FormatInt(m.Delta, 10)
	case GaugeTypeName:
		return strconv.FormatFloat(m.Value, 'f', -1, 64)
	default:
		return ""
	}
}

// The same as Metic, but replaces Delta and Value with pointers to allow for omitempty
type MetricJSON struct {
	*Metric
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

func NewMetricJSON(m *Metric) (mJSON *MetricJSON) {
	mJSON = &MetricJSON{Metric: m}

	switch m.MType {
	case CounterTypeName:
		mJSON.Delta = &m.Delta
	case GaugeTypeName:
		mJSON.Value = &m.Value
	}

	return
}
