package parser

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Parse absolute url address
// If scheme not set set 'http://' as default
func Addr(addr *string) func(string) error {
	return func(flagValue string) error {
		// Address has to have scheme. Add it manually if not set (cause weird tests use that)
		if !strings.Contains(flagValue, "://") {
			flagValue = "http://" + flagValue
		}

		url, err := url.Parse(flagValue)

		if err != nil {
			return err
		}

		*addr = url.String()
		return nil
	}
}

// Parse interval to time.Duration
// The behavior is the same as time.Duration, but if units not specified than 'seconds' is used.
func Interval(interval *time.Duration) func(string) error {
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
		*interval = d
		return nil
	}
}

// Parse level as allowed log level
func LogLevel(level *string) func(string) error {
	levels := []string{"debug", "info", "warning", "error"}
	return func(flagValue string) error {
		value := strings.ToLower(flagValue)
		for _, a := range levels {
			if value == a {
				*level = a
				return nil
			}
		}

		return fmt.Errorf("invalid log level: '%s'\nAllowed (case-insensitive: %s)", flagValue, strings.Join(levels, ", "))
	}
}
