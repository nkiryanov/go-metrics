package opts

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nkiryanov/go-metrics/internal/logger"
)

type Options struct {
	ListenAddr string
	LogLevel   string

	FilePath      string
	StoreInterval time.Duration
	Restore       bool
}

func (opts *Options) Parse() {
	flag.Func("a", "Server listen address in format 'host:port'", parseListenAddr(&opts.ListenAddr))
	flag.Func("i", "Time interval after which server data are saved to file (value 0 makes writing synchronous)", parseStoreInterval(&opts.StoreInterval))
	flag.StringVar(&opts.LogLevel, "l", opts.LogLevel, "Log level like 'info', 'debug', 'error', etc.")
	flag.StringVar(&opts.FilePath, "f", opts.FilePath, "File storage path, like '/tmp/server_data_json.json")
	flag.BoolVar(&opts.Restore, "r", opts.Restore, "Restore initial state from the file storage file")

	// Parse command line args
	flag.Parse()

	// Parse env arguments.
	// Should have precedence if has correct values or ignore if env var is set but values incorrect.
	opts.parseEnv()
}

func (opts *Options) parseEnv() {
	// Helpers to use in envMap with other custom parsers
	parseString := func(optValue *string) func(string) error {
		return func(envValue string) error {
			*optValue = envValue
			return nil
		}
	}

	parseBool := func(optValue *bool) func(string) error {
		return func(envValue string) error {
			value, err := strconv.ParseBool(envValue)
			if err != nil {
				return err
			}
			*optValue = value
			return nil
		}
	}

	envMap := map[string]func(string) error{
		"ADDRESS":           parseListenAddr(&opts.ListenAddr),
		"LOG_LEVEL":         parseString(&opts.LogLevel),
		"FILE_STORAGE_PATH": parseString(&opts.FilePath),
		"RESTORE":           parseBool(&opts.Restore),
	}

	for key, parseFn := range envMap {
		if envVar := os.Getenv(key); envVar != "" {
			if err := parseFn(envVar); err != nil {
				logger.Slog.Errorw("invalid env variable, skipped", key, envVar, "error", err.Error())
			} else {
				logger.Slog.Infow("Set args form env", key, envVar)
			}
		}
	}
}

func parseListenAddr(listenAddr *string) func(string) error {
	return func(flagValue string) error {
		parts := strings.Split(flagValue, ":")

		if len(parts) != 2 {
			return errors.New("need address in a form host:port")
		}

		port, err := strconv.Atoi(parts[1])
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

// Parse storeInterval to time.Duration
// The behavior is the same as time.Duration, but if units not specified than 'seconds' is used.
func parseStoreInterval(intv *time.Duration) func(string) error {
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
			return errors.New("negative values not allowed")
		}
		*intv = d
		return nil
	}
}
