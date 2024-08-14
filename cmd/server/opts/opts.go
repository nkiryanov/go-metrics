package opts

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

type Options struct {
	ListenAddr string
}

func (opts *Options) Parse() {
	flag.Func("a", "server listen address in format 'host:port'", parseListenAddr(&opts.ListenAddr))
	flag.Parse()
}

func parseListenAddr(listenAddr *string) func(string) error {
	return func(flagValue string) error {
		hp := strings.Split(flagValue, ":")

		if len(hp) != 2 {
			return errors.New("need address in a form host:port")
		}

		port, err := strconv.Atoi(hp[1])
		if err != nil {
			return err
		}

		if port < 0 || port > 65535 {
			return errors.New("port has to be in range 0-65535")
		}

		*listenAddr = flagValue
		return nil
	}
}
