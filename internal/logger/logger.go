package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Constants for logging levels
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
)

type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	
	With(args ...any) Logger
	WithGroup(name string) Logger
}

// slogLogger implementation of Logger interface based on slog
type slogLogger struct {
	logger *slog.Logger
}

// parseLevelString converts string level to slog.Level
func parseLevelString(level string) slog.Level {
	switch strings.ToUpper(level) {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:  
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo // default to INFO
	}
}

// NewLogger creates a new logger instance with specified level
func NewLogger(level string) Logger {
	// Configure handler for pretty console output
	opts := &slog.HandlerOptions{
		Level: parseLevelString(level),
		AddSource: true, // adds file and line information
	}
	
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	
	return &slogLogger{logger: logger}
}

// NewJSONLogger creates a logger with JSON format (for production)
func NewJSONLogger(level string) Logger {
	opts := &slog.HandlerOptions{
		Level: parseLevelString(level),
		AddSource: true,
	}
	
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	
	return &slogLogger{logger: logger}
}

// NewNoOpLogger creates a no-op logger using slog.DiscardHandler
func NewNoOpLogger() Logger {
	logger := slog.New(slog.DiscardHandler)
	return &slogLogger{logger: logger}
}

// Logger interface method implementations

func (l *slogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *slogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *slogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *slogLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// With returns a new logger with additional key-value pairs added to all log entries
func (l *slogLogger) With(args ...any) Logger {
	return &slogLogger{logger: l.logger.With(args...)}
}

// WithGroup returns a new logger with all subsequent attributes grouped under the given name
func (l *slogLogger) WithGroup(name string) Logger {
	return &slogLogger{logger: l.logger.WithGroup(name)}
}