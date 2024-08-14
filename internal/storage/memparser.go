package storage

import (
	"fmt"
	"strconv"
)

type MemParser struct{}

func (p MemParser) ParseCounter(s string) (Counter, error) {
	counter, err := strconv.ParseInt(s, 10, 64)
	return Counter(counter), err
}

func (p MemParser) ParseGauge(s string) (Gauge, error) {
	gauge, err := strconv.ParseFloat(s, 64)
	return Gauge(gauge), err
}

func (p MemParser) Parse(mType string, s string) (Storable, error) {
	switch mType {
	case CounterTypeName:
		return p.ParseCounter(s)
	case GaugeTypeName:
		return p.ParseGauge(s)
	default:
		return nil, fmt.Errorf("not supported metric type: %s", mType)
	}
}
