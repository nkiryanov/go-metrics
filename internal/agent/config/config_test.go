package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {

	defaultConfig := Config{
		LogLevel:        "info",
		ReportAddr:      "http://localhost:9090/",
		ReportInterval:  3 * time.Hour,
		SecretKey:       "default-secret-key",
		CollectInterval: 5 * time.Hour,
	}

	tests := []struct {
		name           string
		args           []string
		envVars        map[string]string
		expectedConfig Config
	}{
		{
			name: "env vars takes precedence",
			args: []string{
				"-l", "debug",
				"-a", "http://example.com", // Report Address
				"-r", "3m", // Report Interval
				"-k", "cli-secret-key", // Secret key
				"-p", "5m", // Collect Interval
			},
			envVars: map[string]string{
				"LOG_LEVEL":       "error",
				"ADDRESS":         "127.0.0.1:9090", // Report Address
				"REPORT_INTERVAL": "3s",
				"KEY":             "env-secret-key", // Secret key
				"POLL_INTERVAL":   "5s",
			},
			expectedConfig: Config{
				LogLevel:        "error",
				ReportAddr:      "http://127.0.0.1:9090",
				ReportInterval:  3 * time.Second,
				SecretKey:       "env-secret-key",
				CollectInterval: 5 * time.Second,
			},
		},
		{
			name: "use cli if env empty",
			args: []string{
				"-l", "debug",
				"-a", "http://example.com", // Report Address
				"-r", "3m", // Report Interval
				"-k", "cli-secret-key", // Secret key
				"-p", "5m", // Collect Interval
			},
			envVars: map[string]string{},
			expectedConfig: Config{
				LogLevel:        "debug",
				ReportAddr:      "http://example.com",
				ReportInterval:  3 * time.Minute,
				SecretKey:       "cli-secret-key",
				CollectInterval: 5 * time.Minute,
			},
		},
		{
			name:           "use default if env and cli not set",
			args:           []string{},
			envVars:        map[string]string{},
			expectedConfig: defaultConfig,
		},
		{
			name: "use cli arguments if env value set but invalid",
			args: []string{
				"-r", "5m", // Report interval
			},
			envVars: map[string]string{
				"REPORT_INTERVAL": "-2",
			},
			expectedConfig: Config{
				LogLevel:        defaultConfig.LogLevel,
				ReportAddr:      defaultConfig.ReportAddr,
				ReportInterval:  5 * time.Minute, // loaded form cli
				SecretKey:       defaultConfig.SecretKey,
				CollectInterval: defaultConfig.CollectInterval,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flag.CommandLine
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			// Set environment variables
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Set command line arguments
			origArgs := os.Args
			defer func() { os.Args = origArgs }()
			os.Args = append([]string{os.Args[0]}, tt.args...) // set command-line arguments

			cfg := defaultConfig
			cfg.MustLoad()

			require.Equal(t, tt.expectedConfig, cfg)
		})
	}
}
