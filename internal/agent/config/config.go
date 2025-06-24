package config

import (
	"flag"
	"os"
	"time"

	"github.com/nkiryanov/go-metrics/internal/config/parser"
	"github.com/nkiryanov/go-metrics/internal/logger"
)

type Config struct {
	LogLevel string

	// Agent reporters settings
	ReportAddr      string
	ReportInterval  time.Duration
	ReportRateLimit int // Limit reporter connections to server
	SecretKey       string

	// Agent collectors settings
	CollectInterval time.Duration
}

func (cfg *Config) MustLoad() {
	flag.Func("a", "report address in format http://reports.com", parser.Addr(&cfg.ReportAddr))
	flag.Func("p", "collector polling interval (in seconds by default). Should be positive number like: 10 or 10s or 1m.", parser.Interval(&cfg.CollectInterval))
	flag.Func("r", "report interval (in seconds by default). Should be positive number like: 10 or 10s or 1m.", parser.Interval(&cfg.ReportInterval))
	flag.Func("l", "report rate limit server connections", parser.PositiveInt(&cfg.ReportRateLimit))
	flag.Func("v", "log level like info, debug, etc.", parser.LogLevel(&cfg.LogLevel))

	flag.StringVar(&cfg.SecretKey, "k", cfg.SecretKey, "secret key to sign reported metrics")

	flag.Parse()
	cfg.loadEnvs()
}

// Load variables from environment
// If variable is valid - load it, if not - log error and skip
func (cfg *Config) loadEnvs() {
	lgr := logger.NewLogger("INFO") // there is no logger for config, just use the

	envMap := map[string]func(string) error{
		"LOG_LEVEL":       func(value string) error { cfg.LogLevel = value; return nil },
		"ADDRESS":         parser.Addr(&cfg.ReportAddr), // Report Address
		"REPORT_INTERVAL": parser.Interval(&cfg.ReportInterval),
		"RATE_LIMIT":      parser.PositiveInt(&cfg.ReportRateLimit),
		"KEY":             func(value string) error { cfg.SecretKey = value; return nil },
		"POLL_INTERVAL":   parser.Interval(&cfg.CollectInterval), // Collect Interval
	}

	for key, parseFn := range envMap {
		envVar := os.Getenv(key)
		if envVar != "" {
			err := parseFn(envVar)
			if err != nil {
				lgr.Error("Invalid env variable, skipped", key, envVar, "error", err.Error())
			}
		}
	}
}
