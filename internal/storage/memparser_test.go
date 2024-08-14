package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemParser_ParseGauge(t *testing.T) {
	type expected struct {
		value   Gauge
		isError bool
	}
	tests := []struct {
		name     string
		toParse  string
		expected expected
	}{
		{
			name:     "valid value to parse",
			toParse:  "10.23",
			expected: expected{value: Gauge(10.23), isError: false},
		},
		{
			name:     "invalid value to parse",
			toParse:  "not-valid-gauge",
			expected: expected{value: Gauge(0), isError: true},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := MemParser{}.ParseGauge(tc.toParse)

			if tc.expected.isError {
				require.Error(t, err)
			}
			assert.Equal(t, tc.expected.value, got)
		})
	}
}

func TestMemParser_ParseCounter(t *testing.T) {
	type expected struct {
		value   Counter
		isError bool
	}
	tests := []struct {
		name     string
		toParse  string
		expected expected
	}{
		{
			name:     "valid value to parse",
			toParse:  "10",
			expected: expected{value: Counter(10), isError: false},
		},
		{
			name:     "invalid value to parse",
			toParse:  "10.23", // even float is not valid
			expected: expected{value: Counter(0), isError: true},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := MemParser{}.ParseCounter(tc.toParse)

			if tc.expected.isError {
				require.Error(t, err)
			}
			assert.Equal(t, tc.expected.value, got)
		})
	}
}

func TestMemParser_Parse(t *testing.T) {
	type expected struct {
		value   Storable
		isError bool
	}
	type fnArgs struct {
		mType string
		value string
	}
	tests := []struct {
		name     string
		fnArgs   fnArgs
		expected expected
	}{
		{
			name:     "valid counter",
			fnArgs:   fnArgs{mType: "counter", value: "20"},
			expected: expected{Counter(20), false},
		},
		{
			name:     "invalid counter",
			fnArgs:   fnArgs{mType: "counter", value: "20.123"},
			expected: expected{Counter(0), true},
		},
		{
			name:     "valid gauge",
			fnArgs:   fnArgs{mType: "gauge", value: "20.123"},
			expected: expected{Gauge(20.123), false},
		},
		{
			name:     "invalid gauge",
			fnArgs:   fnArgs{mType: "gauge", value: "not-valid-gauge"},
			expected: expected{Gauge(0.0), true},
		},
		{
			name:     "invalid metric type",
			fnArgs:   fnArgs{mType: "invalid-type", value: "10"},
			expected: expected{nil, true},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := MemParser{}.Parse(tc.fnArgs.mType, tc.fnArgs.value)

			if tc.expected.isError {
				require.Error(t, err)
			}
			assert.Equal(t, tc.expected.value, got)
		})
	}
}
