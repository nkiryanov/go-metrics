package opts

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNetAddress_Set(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected NetAddress
		shouldOK bool
	}{
		{
			name:     "valid localhost, ok",
			input:    "localhost:8029",
			expected: NetAddress{"localhost", 8029},
			shouldOK: true,
		},
		{
			name:     "valid hostname, ok",
			input:    "money.nkiryanov.com:8080",
			expected: NetAddress{"money.nkiryanov.com", 8080},
			shouldOK: true,
		},
		{
			name:     "no port, bad",
			input:    "localhost",
			expected: NetAddress{"default", 1111},
			shouldOK: false,
		},
		{
			name:     "no host, ok",
			input:    ":5423",
			expected: NetAddress{"", 5423},
			shouldOK: true,
		},
		{
			name:     "invalid port, bad",
			input:    "localhost:80000",
			expected: NetAddress{"default", 1111},
			shouldOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := flag.NewFlagSet("test", flag.ContinueOnError)
			flags.SetOutput(io.Discard)
			na := NetAddress{"default", 1111}
			flags.Var(&na, "net-address", "Net address")

			err := flags.Parse([]string{"-net-address", tc.input})

			if tc.shouldOK {
				require.Nil(t, err)
			} else {
				require.Error(t, err)
			}
			assert.Equal(t, tc.expected, na)
		})
	}
}
