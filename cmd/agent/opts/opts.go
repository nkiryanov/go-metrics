package opts

import (
	"errors"
	"flag"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Options struct {
	ReptAddr ReptAddr

	PollIntv IntvValue
	ReptIntv IntvValue
}

func (opts *Options) Parse() {
	flag.Var(&opts.ReptAddr, "a", "report address in format http://reports.com")
	flag.Var(&opts.PollIntv, "p", "capturer polling interval (in seconds by default). Should be positive number like: 10 or 10s or 1m.")
	flag.Var(&opts.ReptIntv, "r", "report interval (in seconds by default). Should be positive number like: 10 or 10s or 1m.")
	flag.Parse()
}

type ReptAddr string

func (ra ReptAddr) String() string {
	return string(ra)
}

func (ra *ReptAddr) Set(s string) error {
	// Add scheme manually if not set. Cause weird tests try that
	if !strings.Contains(s, "://") {
		s = "http://" + s
	}

	// Just us url parser cause it's the case
	url, err := url.Parse(s)

	if err != nil {
		return err
	}

	if url.Scheme == "" || url.Host == "" {
		return errors.New("reporter address has to be in format http://valid-net-address")
	}

	*ra = ReptAddr(url.String())
	return nil
}

// Interval for parsing purpose only.
// Will be converted to duration as soon as parsed
type IntvValue time.Duration

func (i IntvValue) String() string {
	return time.Duration(i).String()
}

func (i *IntvValue) Set(s string) error {
	// If no suffix add 's'
	if _, err := strconv.Atoi(s); err == nil {
		s += "s"
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	if d < 0 {
		return errors.New("must be positive")
	}
	*i = IntvValue(d)
	return nil
}
