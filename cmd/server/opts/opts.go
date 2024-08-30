package opts

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"

	"github.com/nkiryanov/go-metrics/internal/logger"
)

type Options struct {
	ListenAddr string
	LogLevel   string
}

func (opts *Options) Parse() {
	flag.Func("a", "server listen address in format 'host:port'", parseListenAddr(&opts.ListenAddr))
	flag.StringVar(&opts.LogLevel, "l", "info", "log level like info, debug, error, etc.")
	flag.Parse()

	opts.parseEnv()
}

func (opts *Options) parseEnv() {
	envMap := map[string]func(string) error{
		"ADDRESS":   parseListenAddr(&opts.ListenAddr),
		"LOG_LEVEL": func(value string) error { opts.LogLevel = value; return nil },
	}

	for key, parseFn := range envMap {
		if envVar := os.Getenv(key); envVar != "" {
			if err := parseFn(envVar); err != nil {
				logger.Slog.Error("invalid env variable, skipped", key, envVar, "error", err.Error())
			} else {
				logger.Slog.Info("Set args form env", key, envVar)
			}
		}
	}
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
