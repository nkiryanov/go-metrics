package opts

import (
	"errors"
	"flag"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	ReportAddr string
	LogLevel   string

	PollInterval   time.Duration
	ReportInterval time.Duration
	SecretKey      string
}

func (opts *Options) Parse() {
	flag.Func("a", "report address in format http://reports.com", parseReportAddr(&opts.ReportAddr))
	flag.Func("p", "capturer polling interval (in seconds by default). Should be positive number like: 10 or 10s or 1m.", parseInterval(&opts.PollInterval))
	flag.Func("r", "report interval (in seconds by default). Should be positive number like: 10 or 10s or 1m.", parseInterval(&opts.ReportInterval))

	flag.StringVar(&opts.LogLevel, "l", opts.LogLevel, "log level like info, debug, error, etc.")
	flag.StringVar(&opts.SecretKey, "k", opts.SecretKey, "secret key to sign reported metrics")

	flag.Parse()

	opts.parseEnv()
}

func (opts *Options) parseEnv() {
	envMap := map[string]func(string) error{
		"ADDRESS":         parseReportAddr(&opts.ReportAddr),
		"REPORT_INTERVAL": parseInterval(&opts.ReportInterval),
		"POLL_INTERVAL":   parseInterval(&opts.PollInterval),
		"LOG_LEVEL":       func(value string) error { opts.LogLevel = value; return nil },
		"KEY":             func(value string) error { opts.SecretKey = value; return nil },
	}

	for key, parseFn := range envMap {
		if envVar := os.Getenv(key); envVar != "" {
			if err := parseFn(envVar); err != nil {
				slog.Error("invalid env variable, skipped", key, envVar, "error", err.Error())
			} else {
				slog.Info("Set args form env", key, envVar)
			}
		}
	}
}

// Return a func to parse and set value for ReportAddr
func parseReportAddr(ra *string) func(string) error {
	return func(flagValue string) error {
		// ReptAddr has to have scheme. Add it manually if not set (cause weird tests use that)
		if !strings.Contains(flagValue, "://") {
			flagValue = "http://" + flagValue
		}

		url, err := url.Parse(flagValue)

		if err != nil {
			return err
		}

		*ra = url.String()
		return nil
	}
}

// Parse interval to time.Duration
// The behavior is the same as time.Duration, but if units not specified than 'seconds' is used.
func parseInterval(intv *time.Duration) func(string) error {
	return func(flagValue string) error {
		// If no suffix add 's'
		if _, err := strconv.Atoi(flagValue); err == nil {
			flagValue += "s"
		}

		d, err := time.ParseDuration(flagValue)
		if err != nil {
			return err
		}
		if d <= 0 {
			return errors.New("must be positive")
		}
		*intv = d
		return nil
	}
}
