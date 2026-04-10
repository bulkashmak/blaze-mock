package blaze

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// LogOutput controls where the server writes log messages.
type LogOutput int

const (
	// LogNone disables all logging (useful for load testing).
	LogNone LogOutput = iota
	// LogStdout writes logs to standard output with pretty formatting.
	LogStdout
	// LogFile writes logs to a file specified by WithLogFile.
	LogFile
	// LogBoth writes logs to both stdout (pretty) and a file (structured).
	LogBoth
)

// newLogger creates a *slog.Logger based on the logging configuration.
// The caller is responsible for closing the returned file (if any).
func newLogger(output LogOutput, filePath string) (*slog.Logger, *os.File, error) {
	switch output {
	case LogNone:
		return slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil

	case LogStdout:
		return slog.New(&prettyHandler{w: os.Stdout}), nil, nil

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
		fileHandler := slog.NewTextHandler(f, nil)
		return slog.New(&multiHandler{
			handlers: []slog.Handler{
				&prettyHandler{w: os.Stdout},
				fileHandler,
			},
		}), f, nil

	default:
		return slog.New(&prettyHandler{w: os.Stdout}), nil, nil
	}
}

// ANSI color codes.
const (
	colorReset  = "\033[0m"
	colorDim    = "\033[2m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorRed    = "\033[31m"
)

// prettyHandler formats log records for human-readable stdout output with colors.
type prettyHandler struct {
	w io.Writer
}

func (h *prettyHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *prettyHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }

func (h *prettyHandler) WithGroup(_ string) slog.Handler { return h }

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder

	ts := r.Time.Format(time.TimeOnly)
	levelColor := colorGreen
	switch r.Level {
	case slog.LevelError:
		levelColor = colorRed
	case slog.LevelWarn:
		levelColor = colorYellow
	}

	b.WriteString(fmt.Sprintf("%s%s%s %s%-5s%s %s%s%s\n",
		colorDim, ts, colorReset,
		levelColor, r.Level.String(), colorReset,
		colorCyan, r.Message, colorReset,
	))

	r.Attrs(func(a slog.Attr) bool {
		h.writeAttr(&b, a, "  ")
		return true
	})

	_, err := fmt.Fprint(h.w, b.String())
	return err
}

func (h *prettyHandler) writeAttr(b *strings.Builder, a slog.Attr, indent string) {
	switch a.Value.Kind() {
	case slog.KindGroup:
		b.WriteString(fmt.Sprintf("%s%s%s%s:\n", indent, colorDim, a.Key, colorReset))
		for _, ga := range a.Value.Group() {
			h.writeAttr(b, ga, indent+"  ")
		}
	default:
		b.WriteString(fmt.Sprintf("%s%s%s%s: %s\n", indent, colorDim, a.Key, colorReset, a.Value.String()))
	}
}

// multiHandler fans out to multiple slog.Handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if err := h.Handle(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}
