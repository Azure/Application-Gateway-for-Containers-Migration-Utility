// Package logging provides a logger instance configured based on whether the process is running in an interactive terminal.
package logging

import (
	"os"
	"time"

	"log/slog"

	"github.com/lmittmann/tint"

	"golang.org/x/term"
)

func New(_ string) *slog.Logger {
	if term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) {
		return slog.New(tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}))
	}

	return slog.New(slog.NewJSONHandler(os.Stderr, nil))
}
