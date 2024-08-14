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
	ReptAddr string

	PollIntv time.Duration
	ReptIntv time.Duration
}

func (opts *Options) Parse() {
	flag.Func("a", "report address in format http://reports.com", parseReptAddr(&opts.ReptAddr))
	flag.Func("p", "capturer polling interval (in seconds by default). Should be positive number like: 10 or 10s or 1m.", parseIntv(&opts.PollIntv))
	flag.Func("r", "report interval (in seconds by default). Should be positive number like: 10 or 10s or 1m.", parseIntv(&opts.ReptIntv))
	flag.Parse()

	opts.parseEnv()
}

func (opts *Options) parseEnv() {
	envMap := map[string]func(string) error{
		"ADDRESS":         parseReptAddr(&opts.ReptAddr),
		"REPORT_INTERVAL": parseIntv(&opts.ReptIntv),
		"POLL_INTERVAL":   parseIntv(&opts.PollIntv),
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
func parseReptAddr(ra *string) func(string) error {
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

// Interval for parsing purpose only.
// Will be converted to duration as soon as parsed
type IntvValue time.Duration

func parseIntv(intv *time.Duration) func(string) error {
	return func(flagValue string) error {
		// If no suffix add 's'
		if _, err := strconv.Atoi(flagValue); err == nil {
			flagValue += "s"
		}

		d, err := time.ParseDuration(flagValue)
		if err != nil {
			return err
		}
		if d < 0 {
			return errors.New("must be positive")
		}
		*intv = d
		return nil
	}
}
