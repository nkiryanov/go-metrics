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
	ReptAddr string

	PollIntv   time.Duration
	ReptIntv time.Duration
}

func ParseOptions() *Options {
	opts := &Options{
		PollIntv:   PollInterval,
		ReptIntv: ReportInterval,
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

	flag.StringVar(&opts.ReptAddr, "a", "http://localhost:8080", "server report address in format 'scheme://host:port'")
	flag.Func("p", "polling interval in seconds", parseDurationFunc(&opts.PollIntv))
	flag.Func("r", "reporting interval in seconds", parseDurationFunc(&opts.ReptIntv))

	flag.Parse()

	if !strings.Contains(opts.ReptAddr, "://") {
		opts.ReptAddr = "http://" + opts.ReptAddr
	}

	return opts
}
