package blaze

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

// LogOutput controls where the server writes log messages.
type LogOutput int

const (
	// LogNone disables all logging (useful for load testing).
	LogNone LogOutput = iota
	// LogStdout writes logs to standard output.
	LogStdout
	// LogFile writes logs to a file specified by WithLogFile.
	LogFile
	// LogBoth writes logs to both stdout and a file.
	LogBoth
)

// newLogger creates a *slog.Logger based on the logging configuration.
// The caller is responsible for closing the returned file (if any).
func newLogger(output LogOutput, filePath string) (*slog.Logger, *os.File, error) {
	switch output {
	case LogNone:
		return slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil

	case LogStdout:
		return slog.New(slog.NewTextHandler(os.Stdout, nil)), nil, nil

	case LogFile:
		if filePath == "" {
			return nil, nil, fmt.Errorf("blaze: LogFile requires WithLogFile option")
		}
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("blaze: failed to open log file %s: %w", filePath, err)
		}
		return slog.New(slog.NewTextHandler(f, nil)), f, nil

	case LogBoth:
		if filePath == "" {
			return nil, nil, fmt.Errorf("blaze: LogBoth requires WithLogFile option")
		}
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, nil, fmt.Errorf("blaze: failed to open log file %s: %w", filePath, err)
		}
		w := io.MultiWriter(os.Stdout, f)
		return slog.New(slog.NewTextHandler(w, nil)), f, nil

	default:
		return slog.New(slog.NewTextHandler(os.Stdout, nil)), nil, nil
	}
}
