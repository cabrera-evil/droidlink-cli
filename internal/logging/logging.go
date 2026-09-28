// Package logging provides droidlink's slog setup. Loggers are never stored
// in package-level globals (per the repo's sloglint no-global rule); a
// logger is built once in internal/cli's root command and threaded through
// context.Context for the lifetime of a single invocation.
package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"

	"droidlink/internal/utils"
)

type contextKey struct{}

// Options configures logger construction.
type Options struct {
	Level   slog.Level
	NoColor bool
	Writer  io.Writer // defaults to os.Stderr when nil
}

// New builds a slog.Logger writing to Options.Writer (default os.Stderr),
// colorized via tint when the destination is an interactive terminal and
// NoColor is false.
func New(opts Options) *slog.Logger {
	w := opts.Writer
	if w == nil {
		w = os.Stderr
	}

	if !opts.NoColor {
		if f, ok := w.(*os.File); ok && utils.IsInteractive(f) {
			return slog.New(tint.NewTextHandler(w, &tint.Options{Level: opts.Level}))
		}
	}

	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: opts.Level}))
}

// ParseLevel parses a case-insensitive log level name (debug|info|warn|error).
func ParseLevel(name string) (slog.Level, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(name)); err != nil {
		return 0, fmt.Errorf("invalid log level %q: %w", name, err)
	}

	return level, nil
}

// WithLogger returns a copy of ctx carrying logger.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// FromContext returns the logger stored in ctx, or a discard logger if none
// was attached.
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(contextKey{}).(*slog.Logger); ok {
		return logger
	}

	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
