package opts

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

type Options struct {
	ListenAddr NetAddress
}

func (o *Options) Parse() {
	flag.Var(&o.ListenAddr, "a", "server listen address in format 'host:port'")
	flag.Parse()
}

type NetAddress struct {
	Host string
	Port int
}

func (a NetAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *NetAddress) Set(s string) error {
	hp := strings.Split(s, ":")

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

	a.Host = hp[0]
	a.Port = port

	return nil
}
