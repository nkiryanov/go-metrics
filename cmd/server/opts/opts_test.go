package opts

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNetAddress_Set(t *testing.T) {
	var na NetAddress

	tests := []struct {
		name     string
		input    string
		shouldOK bool
	}{
		{
			name:     "valid localhost, ok",
			input:    "localhost:8029",
			shouldOK: true,
		},
		{
			name:     "valid hostname, ok",
			input:    "money.nkiryanov.com:8080",
			shouldOK: true,
		},
		{
			name:     "no port, bad",
			input:    "localhost",
			shouldOK: false,
		},
		{
			name:     "no host, ok",
			input:    ":5423",
			shouldOK: true,
		},
		{
			name:     "invalid port, bad",
			input:    "localhost:80000",
			shouldOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := flag.NewFlagSet("test", flag.ContinueOnError)
			flags.SetOutput(io.Discard)
			flags.Var(&na, "net-address", "Net address")

			err := flags.Parse([]string{"-net-address", tc.input})

			if tc.shouldOK {
				require.Nil(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
