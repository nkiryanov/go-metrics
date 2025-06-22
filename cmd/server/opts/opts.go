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

	// In memory database with persistent storage at file
	DataFilePath   string
	SaveInterval   time.Duration
	RestoreOnStart bool

	// If set postgres storage will be used
	DatabaseDsn string

	// Secret key to verify hmac of reported metrics
	SecretKey string
}

func (opts *Options) Parse() {
	flag.Func("a", "Server listen address in format 'host:port'", parseListenAddr(&opts.ListenAddr))
	flag.Func("i", "Time interval after which server data are saved to file (value 0 makes writing synchronous)", parseSaveInterval(&opts.SaveInterval))
	flag.StringVar(&opts.LogLevel, "l", opts.LogLevel, "Log level like 'info', 'debug', 'error', etc.")
	flag.StringVar(&opts.DataFilePath, "f", opts.DataFilePath, "File storage path, like '/tmp/server_data_json.json")
	flag.StringVar(&opts.DatabaseDsn, "d", opts.DatabaseDsn, "Database connection string like 'postgres://user:password@localhost:5432/dbname'")
	flag.StringVar(&opts.SecretKey, "k", opts.SecretKey, "Secret Key to verify HMAC (provided in HashSHA256 header) of reporting metrics. Takes no effect if not set or empty")
	flag.BoolVar(&opts.RestoreOnStart, "r", opts.RestoreOnStart, "Restore initial state from the file storage file on server start")

	// Parse command line args
	flag.Parse()

	// Parse env arguments.
	// Should have precedence if has correct values or ignore if env var is set but values incorrect.
	opts.parseEnv()
}

func (opts *Options) parseEnv() {
	fallbackLgr := logger.NewLogger("INFO") // there is no logger for config, just use fallback one

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
		"FILE_STORAGE_PATH": parseString(&opts.DataFilePath),
		"RESTORE":           parseBool(&opts.RestoreOnStart),
		"DATABASE_DSN":      parseString(&opts.DatabaseDsn),
		"KEY":               parseString(&opts.SecretKey),
	}

	for key, parseFn := range envMap {
		if envVar := os.Getenv(key); envVar != "" {
			if err := parseFn(envVar); err != nil {
				fallbackLgr.Error("invalid env variable, skipped", key, envVar, "error", err.Error())
			} else {
				fallbackLgr.Info("Set args form env", key, envVar)
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

// Parse saveInterval to time.Duration
// The behavior is the same as time.Duration, but if units not specified than 'seconds' is used.
func parseSaveInterval(intv *time.Duration) func(string) error {
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
