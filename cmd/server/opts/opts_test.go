package opts

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseListenAddr(t *testing.T) {
	defaultListenAddr := "default:1111"
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
			expected: defaultListenAddr,
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
			expected: defaultListenAddr,
			shouldOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			listenAddr := defaultListenAddr
			parseFn := parseListenAddr(&listenAddr)

			err := parseFn(tc.input)

			if tc.shouldOK {
				require.Nil(t, err)
			} else {
				require.Error(t, err)
			}
			assert.Equal(t, tc.expected, listenAddr)
		})
	}
}

func Test_parseStoreInterval(t *testing.T) {
	t.Run("valid valued", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected time.Duration
		}{
			{
				"time duration ok",
				"10m",
				10 * time.Minute,
			},
			{
				"int parsed as seconds",
				"5",
				5 * time.Second,
			},
			{
				"zero allowed",
				"0",
				0,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				storeInterval := 300 * time.Second
				parseFn := parseStoreInterval(&storeInterval)

				err := parseFn(tc.input)

				require.Nil(t, err)
				assert.Equal(t, tc.expected, storeInterval)
			})
		}
	})

	t.Run("negative not allowed", func(t *testing.T) {
		storeInterval := 300 * time.Second
		parseFn := parseStoreInterval(&storeInterval)

		err := parseFn("-23s")

		require.Error(t, err)
		require.Equal(t, 300*time.Second, storeInterval, "initial value has not be changed")
	})
}

func TestOptions(t *testing.T) {
	defaultOpts := Options{
		ListenAddr:    "localhost:8080",
		LogLevel:      "info",
		FilePath:      "/tmp/default_data.json",
		StoreInterval: 300 * time.Second,
		Restore:       false,
	}

	tests := []struct {
		name            string
		args            []string
		envVars         map[string]string
		expectedOptions Options
	}{
		{
			name: "env vars takes precedence",
			args: []string{"-a", "localhost:1234", "-l", "debug", "-i", "4m", "-f", "/tmp/test.json", "-r", "false"},
			envVars: map[string]string{
				"ADDRESS":           "127.0.0.1:9090",
				"LOG_LEVEL":         "error",
				"FILE_STORAGE_PATH": "/tmp/env_test.json",
				"STORE_INTERVAL":    "1m",
				"RESTORE":           "TRUE",
			},
			expectedOptions: Options{
				ListenAddr:    "127.0.0.1:9090",
				LogLevel:      "error",
				FilePath:      "/tmp/env_test.json",
				StoreInterval: 1 * time.Minute,
			},
		},
		{
			name:    "use cli arguments if set",
			args:    []string{"-a", "localhost:1234", "-l", "debug", "-i", "4m", "-f", "/tmp/test.json", "-r", "true"},
			envVars: map[string]string{},
			expectedOptions: Options{
				ListenAddr:    "localhost:1234",
				LogLevel:      "debug",
				StoreInterval: 4 * time.Minute,
				FilePath:      "/tmp/test.json",
				Restore:       true,
			},
		},
		{
			name:            "use default arguments if nothing set",
			args:            []string{},
			envVars:         map[string]string{},
			expectedOptions: defaultOpts,
		},
		{
			name:    "use cli arguments if env value set but invalid",
			args:    []string{"-i", "1m"},
			envVars: map[string]string{"STORE_INTERVAL": "-2"}, // Invalid store interval
			expectedOptions: Options{
				ListenAddr:    defaultOpts.ListenAddr,
				LogLevel:      defaultOpts.LogLevel,
				FilePath:      defaultOpts.FilePath,
				StoreInterval: 1 * time.Minute, // Should use cli argument cause env variable invalid
				Restore:       defaultOpts.Restore,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flag.CommandLine
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Set (and clean on exit) environment variables
			for key, value := range tc.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Set command-line args
			os.Args = append([]string{os.Args[0]}, tc.args...)

			opts := defaultOpts

			opts.Parse()

			assert.Equal(t, tc.expectedOptions.ListenAddr, opts.ListenAddr)
			assert.Equal(t, tc.expectedOptions.LogLevel, opts.LogLevel)
			assert.Equal(t, tc.expectedOptions.FilePath, opts.FilePath)
		})
	}
}
