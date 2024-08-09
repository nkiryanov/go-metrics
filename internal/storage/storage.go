package storage

//go:generate moq -out mocks/storage.go -pkg mocks -skip-ensure -fmt goimports . Storage

type IterFunc func(mType string, mName string, mValue Storable)

type Storage interface {
	GetCounter(mName string) (Counter, bool)
	UpdateCounter(mName string, value Counter) Counter
	IterateCounters(func(mName string, value Counter))

	GetGauge(mName string) (Gauge, bool)
	UpdateGauge(mName string, value Gauge) Gauge
	IterateGauges(func(mName string, value Gauge))

	// Polymorphic methods
	Len() int
	GetMetric(mType string, mName string) (Storable, bool, error)
	UpdateMetric(mName string, mValue Storable) (Storable, error)
	Iterate(IterFunc)
}
