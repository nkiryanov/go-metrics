package opts

import (
	"flag"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReptAddr_Set(t *testing.T) {
	defaultRa := ReptAddr("https://default.com/update")

	tests := []struct {
		name     string
		input    string
		expected ReptAddr
		shouldOk bool
	}{
		{
			name:     "valid url, ok",
			input:    "http://go-metrics.com",
			expected: "http://go-metrics.com",
			shouldOk: true,
		},
		{
			name:     "valid url with port, ok",
			input:    "http://go-metrics.com:8999/update",
			expected: "http://go-metrics.com:8999/update",
			shouldOk: true,
		},
		{
			name:     "add http prefix, ok",
			input:    "localhost:8080",
			expected: "http://localhost:8080",
			shouldOk: true,
		},
		{
			name:     "invalid url, bad",
			input:    "http://       ya.ru",
			expected: defaultRa,
			shouldOk: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := flag.NewFlagSet("test", flag.ContinueOnError)
			flags.SetOutput(io.Discard)
			ra := defaultRa
			flags.Var(&ra, "report-addr", "report address should like https://repor.com/update")

			err := flags.Parse([]string{"-report-addr", tc.input})

			if tc.shouldOk {
				require.Nil(t, err)
			} else {
				require.Error(t, err)
			}
			assert.EqualValues(t, tc.expected, ra)
		})
	}
}

func TestIntvValue_Set(t *testing.T) {
	defaultIntv := 300 * time.Second

	tests := []struct {
		name     string
		input    string
		expected time.Duration
		shouldOk bool
	}{
		{
			name:     "positive int, ok",
			input:    "10",
			expected: 10 * time.Second,
			shouldOk: true,
		},
		{
			name:     "negative int, bad",
			input:    "-10",
			expected: defaultIntv,
			shouldOk: false,
		},
		{
			name:     "seconds ok too, ok",
			input:    "4s",
			expected: 4 * time.Second,
			shouldOk: true,
		},
		{
			name:     "minute ok too, ok",
			input:    "1m",
			expected: 1 * time.Minute,
			shouldOk: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			flags := flag.NewFlagSet("test", flag.ContinueOnError)
			flags.SetOutput(io.Discard)
			intv := IntvValue(defaultIntv)
			flags.Var(&intv, "interval", "interval like duration or just not negative num")

			err := flags.Parse([]string{"-interval", tc.input})

			if tc.shouldOk {
				require.Nil(t, err)
			} else {
				require.Error(t, err)
			}
			assert.EqualValues(t, tc.expected, intv)
		})
	}
}
