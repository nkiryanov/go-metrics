package opts

import (
	"flag"
)

type Options struct {
	ListenAddr string
}

func NewOptions() *Options {
	opts := &Options{}

	flag.StringVar(&opts.ListenAddr, "a", "localhost:8080", "server listen address in format 'host:port'")
	flag.Parse()

	return opts
}
