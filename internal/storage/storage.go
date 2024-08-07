package storage


type Storage interface {
	GetCounter(mName string) (Counter, bool)
	UpdateCounter(mName string, value Counter) Counter
	IterateCounters(func(mName string, value Counter))

	GetGauge(mName string) (Gauge, bool)
	UpdateGauge(mName string, value Gauge) Gauge
	IterateGauges(func(mName string, value Gauge))

	// Polymorphic methods
	GetMetric(mType string, mName string) (Storable, bool, error)
	UpdateMetric(mType string, mName string, mValue Storable) (Storable, error)
	Iterate(func(mType string, mName string, mValue Storable))
}
