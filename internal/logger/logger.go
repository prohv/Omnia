package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// LogLevel represents custom logging modes for Omnia.
type LogMode string

const (
	ModeDefault LogMode = "default"
	ModeVerbose LogMode = "verbose"
	ModeQuiet   LogMode = "quiet"
	ModeJSON    LogMode = "json"
)

// SetupLogger initializes the global slog.Logger according to mode and output writer.
func SetupLogger(mode string, w io.Writer) *slog.Logger {
	if w == nil {
		w = os.Stdout
	}

	normalizedMode := LogMode(strings.ToLower(strings.TrimSpace(mode)))

	var level slog.Level
	var handler slog.Handler

	switch normalizedMode {
	case ModeVerbose:
		level = slog.LevelDebug
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	case ModeQuiet:
		level = slog.LevelError
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	case ModeJSON:
		level = slog.LevelInfo
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	case ModeDefault:
		fallthrough
	default:
		level = slog.LevelInfo
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	}

	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}
