package storage

import (
	"strconv"
)

const (
	CounterTypeName string = "counter"
	GaugeTypeName   string = "gauge"
)

type (
	Gaugeable float64
	Countable int64
)

// Common interface for types storable in storage
type Storable interface {
	String() string
}

type Storage interface {
	GetCounter(mName string) (Countable, bool)
	UpdateCounter(mName string, value Countable) Countable
	IterateCounters(func(mName string, value Countable))

	GetGauge(mName string) (Gaugeable, bool)
	UpdateGauge(mName string, value Gaugeable) Gaugeable
	IterateGauges(func(mName string, value Gaugeable))

	// Polymorphic methods
	UpdateValue(mType string, mName string, mValue string) (Storable, error)
	Iterate(func(mType string, mName string, mValue Storable))
}

func (c Countable) String() string {
	return strconv.FormatInt(int64(c), 10)
}

func (g Gaugeable) String() string {
	return strconv.FormatFloat(float64(g), 'f', -1, 64)
}
