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

func Test_parseSaveInterval(t *testing.T) {
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
				parseFn := parseSaveInterval(&storeInterval)

				err := parseFn(tc.input)

				require.Nil(t, err)
				assert.Equal(t, tc.expected, storeInterval)
			})
		}
	})

	t.Run("negative not allowed", func(t *testing.T) {
		storeInterval := 300 * time.Second
		parseFn := parseSaveInterval(&storeInterval)

		err := parseFn("-23s")

		require.Error(t, err)
		require.Equal(t, 300*time.Second, storeInterval, "initial value has not be changed")
	})
}

func TestOptions(t *testing.T) {
	defaultOpts := Options{
		ListenAddr:     "localhost:8080",
		LogLevel:       "info",
		DataFilePath:   "/tmp/default_data.json",
		SaveInterval:   300 * time.Second,
		RestoreOnStart: false,
		DatabaseDsn:    "postgres://go-metrics@localhost:5432/go-metrics",
		SecretKey:      "",
		PprofAddr:      "",
	}

	tests := []struct {
		name            string
		args            []string
		envVars         map[string]string
		expectedOptions Options
	}{
		{
			name: "env vars takes precedence",
			args: []string{
				"-a", "localhost:1234",
				"-l", "debug",
				"-i", "4m",
				"-f", "/tmp/test.json",
				"-r", "false",
				"-d", "postgres://test@test:5432/test",
				"-k", "cli-secret-key",
				"--pprof", "localhost:9999",
			},
			envVars: map[string]string{
				"ADDRESS":           "127.0.0.1:9090",
				"LOG_LEVEL":         "error",
				"FILE_STORAGE_PATH": "/tmp/env_test.json",
				"STORE_INTERVAL":    "1m",
				"RESTORE":           "TRUE",
				"DATABASE_DSN":      "postgres://envuser@localhost:5432/envdb",
				"KEY":               "env-secret-key",
				"PPROF_ADDRESS":     "127.0.0.1:5050",
			},
			expectedOptions: Options{
				ListenAddr:   "127.0.0.1:9090",
				LogLevel:     "error",
				DataFilePath: "/tmp/env_test.json",
				SaveInterval: 1 * time.Minute,
				DatabaseDsn:  "postgres://envuser@localhost:5432/envdb",
				SecretKey:    "env-secret-key",
				PprofAddr:    "127.0.0.1:5050",
			},
		},
		{
			name: "use cli arguments if set",
			args: []string{
				"-a", "localhost:1234",
				"-l", "debug",
				"-i", "4m",
				"-f", "/tmp/test.json",
				"-r", "true",
				"-d", "postgres://test@test:5432/test",
				"-k", "cli-secret-key",
				"--pprof", "localhost:9999", // using two dashes here cause flag's length more than one letter, but '-pprof' also works
			},
			envVars: map[string]string{},
			expectedOptions: Options{
				ListenAddr:     "localhost:1234",
				LogLevel:       "debug",
				SaveInterval:   4 * time.Minute,
				DataFilePath:   "/tmp/test.json",
				RestoreOnStart: true,
				DatabaseDsn:    "postgres://test@test:5432/test",
				SecretKey:      "cli-secret-key",
				PprofAddr:      "127.0.0.1:9999",
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
				ListenAddr:     defaultOpts.ListenAddr,
				LogLevel:       defaultOpts.LogLevel,
				DataFilePath:   defaultOpts.DataFilePath,
				SaveInterval:   1 * time.Minute, // Should use cli argument cause env variable invalid
				RestoreOnStart: defaultOpts.RestoreOnStart,
				SecretKey:      defaultOpts.SecretKey,
				PprofAddr:      defaultOpts.PprofAddr,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flag.CommandLine
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Set (and clean on exit) environment variables
			for key, value := range tc.envVars {
				_ = os.Setenv(key, value)
				defer os.Unsetenv(key) // nolint:errcheck
			}

			// Set command-line args
			os.Args = append([]string{os.Args[0]}, tc.args...)

			opts := defaultOpts

			opts.Parse()

			assert.Equal(t, tc.expectedOptions.ListenAddr, opts.ListenAddr)
			assert.Equal(t, tc.expectedOptions.LogLevel, opts.LogLevel)
			assert.Equal(t, tc.expectedOptions.DataFilePath, opts.DataFilePath)
		})
	}
}
