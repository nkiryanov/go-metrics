package opts

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	PollInterval   = 2 * time.Second
	ReportInterval = 10 * time.Second

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
		ReportInterval: ReportInterval,
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

	if !strings.Contains(opts.ReportAddr, "://") {
		opts.ReportAddr = "http://" + opts.ReportAddr
	}

	return opts
}
