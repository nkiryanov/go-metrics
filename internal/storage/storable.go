package storage

import (
	"fmt"
	"strconv"
)

const (
	CounterTypeName = "counter"
	GaugeTypeName = "gauge"
)

type (
	Counter int64
	Gauge float64
)

// Common interface for types storable in storage
type Storable interface {
	String() string
}

func ParseStorable(mType string, s string) (Storable, error) {
	switch mType {
	case CounterTypeName:
		return ParseCounter(s)
	case GaugeTypeName:
		return ParseGauge(s)
	default:
		return nil, fmt.Errorf("not supported metric type: %s", mType)
	}
}

func ParseCounter(s string) (Counter, error) {
	counter, err := strconv.ParseInt(s, 10, 64)
	return Counter(counter), err
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func ParseGauge(s string) (Gauge, error) {
	gauge, err := strconv.ParseFloat(s, 64)
	return Gauge(gauge), err
}

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

