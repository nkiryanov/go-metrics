package opts

import (
	"flag"
	"fmt"
	// "log/slog"
	"strconv"
	"time"
)

const (
	PollInterval = 2 * time.Second
	PubInterval  = 10 * time.Second

	PubAddr = "http://localhost:8080"
)

type Options struct {
	ReportAddr string

	PollInterval   time.Duration
	ReportInterval time.Duration
}

func ParseOptions() *Options {
	opts := &Options{
		PollInterval:   PollInterval,
		ReportInterval: PubInterval,
	}

	parseDurationFunc := func(d *time.Duration) func(string) error {
		return func(s string) error {
			seconds, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("invalid duration: %v", err)
			}
			if seconds <= 0 {
				return fmt.Errorf("duration must be greater than 0")
			}
			*d = time.Second * time.Duration(seconds)
			return nil
		}
	}

	flag.StringVar(&opts.ReportAddr, "a", "http://localhost:8080", "server report address in format 'scheme://host:port'")
	flag.Func("p", "polling interval in seconds", parseDurationFunc(&opts.PollInterval))
	flag.Func("r", "reporting interval in seconds", parseDurationFunc(&opts.ReportInterval))

	flag.Parse()

	flag.Parse()

	return opts
}
