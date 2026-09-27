package exporter

import (
	"io"
	"log/slog"
)

// NewLogger builds the exporter logger, selecting debug-level logging when
// debug is true and info-level logging otherwise.
func NewLogger(debug bool, w io.Writer) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
}
