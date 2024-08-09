package storage

import (
	"strconv"
)

const (
	CounterTypeName = "counter"
	GaugeTypeName   = "gauge"
)

type (
	Counter int64
	Gauge   float64
)

//go:generate moq -out mocks/storable.go -pkg mocks -skip-ensure -fmt goimports . Storable
//go:generate moq -out mocks/parser.go -pkg mocks -skip-ensure -fmt goimports . StorableParser

// Common interface for types storable in storage
type Storable interface {
	String() string
	Type() string
}

type StorableParser interface {
	ParseCounter(string) (Counter, error)
	ParseGauge(string) (Gauge, error)
	Parse(mType string, s string) (Storable, error)
}

func (c Counter) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func (c Counter) Type() string {
	return CounterTypeName
}

func (g Gauge) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}

func (g Gauge) Type() string {
	return GaugeTypeName
}
