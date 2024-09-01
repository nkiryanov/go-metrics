package models

import (
	"strconv"
)

const (
	CounterTypeName = "counter"
	GaugeTypeName   = "gauge"
)

type Metric struct {
	ID    string
	MType string
	Delta int64
	Value float64
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
