package opts

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetAddress_Set(t *testing.T) {
	defaultNa := "default:1111"
	tests := []struct {
		name     string
		input    string
		expected string
		shouldOK bool
	}{
		{
			name:     "valid localhost, ok",
			input:    "localhost:8029",
			expected: "localhost:8029",
			shouldOK: true,
		},
		{
			name:     "valid hostname, ok",
			input:    "money.nkiryanov.com:8080",
			expected: "money.nkiryanov.com:8080",
			shouldOK: true,
		},
		{
			name:     "no port, bad",
			input:    "localhost",
			expected: defaultNa,
			shouldOK: false,
		},
		{
			name:     "no host, ok",
			input:    ":5423",
			expected: ":5423",
			shouldOK: true,
		},
		{
			name:     "invalid port, bad",
			input:    "localhost:80000",
			expected: defaultNa,
			shouldOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			na := defaultNa
			parseFn := parseListenAddr(&na)

			err := parseFn(tc.input)

			if tc.shouldOK {
				require.Nil(t, err)
			} else {
				require.Error(t, err)
			}
			assert.Equal(t, tc.expected, na)
		})
	}
}
