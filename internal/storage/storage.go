package storage

import (
	"fmt"
)

//go:generate moq -out mocks/storage.go -pkg mocks -skip-ensure -fmt goimports . Storage
//go:generate moq -out mocks/storable.go -pkg mocks -skip-ensure -fmt goimports . Storable
//go:generate moq -out mocks/parser.go -pkg mocks -skip-ensure -fmt goimports . StorableParser

// Common interface for types storable in storage
type Storable interface {
	fmt.Stringer
	Type() string
}

type StorableParser interface {
	Parse(mType string, s string) (Storable, error)
}

type IterFunc func(mType string, mName string, mValue Storable)

type Storage interface {
	Len() int

	// Get metric from storage
	// If metric type is supported by the storage but 'mName' not found, then the 'ok' bool will be false.
	// If metric type is not supported by the storage, then 'err' will not be nil.
	GetMetric(mType string, mName string) (value Storable, ok bool, err error)

	// Update metric in storage
	// May be implementation specific and not support all the types. In that case should return 'err'.
	UpdateMetric(mName string, mValue Storable) (value Storable, err error)

	// Iterate over stored values with 'iter' func.
	Iterate(iter IterFunc)
}
