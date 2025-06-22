package parser

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Addr(t *testing.T) {
	t.Run("ok cases", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			{
				name:     "valid url",
				input:    "http://go-metrics.com",
				expected: "http://go-metrics.com",
			},
			{
				name:     "valid url with port",
				input:    "http://go-metrics.com:8999/update",
				expected: "http://go-metrics.com:8999/update",
			},
			{
				name:     "add http prefix",
				input:    "localhost:8080",
				expected: "http://localhost:8080",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				addr := ""
				parseFn := Addr(&addr)

				err := parseFn(tc.input)

				require.Nil(t, err)
				assert.EqualValues(t, tc.expected, addr)
			})
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		addr := "https://pornhub.com"
		parseFn := Addr(&addr)

		err := parseFn("http://  ya.ru")

		require.Error(t, err)
		require.Equal(t, "https://pornhub.com", addr, "Address has not be changed on error")
	})
}

func Test_Interval(t *testing.T) {
	t.Run("ok cases", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected time.Duration
		}{
			{
				name:     "positive int",
				input:    "10",
				expected: 10 * time.Second,
			},
			{
				name:     "seconds suffix",
				input:    "4s",
				expected: 4 * time.Second,
			},
			{
				name:     "minute suffix",
				input:    "1m",
				expected: 1 * time.Minute,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				interval := time.Hour
				parseFn := Interval(&interval)

				err := parseFn(tc.input)

				require.Nil(t, err)
				assert.Equal(t, tc.expected, interval)
			})
		}
	})

	t.Run("invalid interval", func(t *testing.T) {
		interval := time.Hour
		parseFn := Interval(&interval)

		err := parseFn("-10")

		require.Error(t, err)
		require.Equal(t, time.Hour, interval, "Initial value must not change")
	})
}

func Test_LogLevel(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
		}{
			{"debug level", "debug", "debug"},
			{"info level", "info", "info"},
			{"warning level", "warning", "warning"},
			{"error level", "error", "error"},
			{"case insensitive", "InFo", "info"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				level := ""
				parseFn := LogLevel(&level)

				err := parseFn(tt.input)

				require.Nil(t, err)
				assert.Equal(t, tt.expected, level)
			})
		}
	})

	t.Run("not expected level fail", func(t *testing.T) {
		level := "info"
		parseFn := LogLevel(&level)

		err := parseFn("fun")

		require.Error(t, err)
		require.Equal(t, "info", level, "level should not change if parsing fail")
	})
}
