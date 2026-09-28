// Package utils holds small shared helpers used across droidlink's internal
// packages.
package utils

import (
	"os"

	"github.com/chzyer/readline"
)

// IsInteractive reports whether f refers to an interactive terminal, used to
// decide between interactive prompts and flag-driven fallbacks. A plain
// os.ModeCharDevice check is insufficient: /dev/null and other non-terminal
// character devices satisfy it too, so this does a real ioctl-based terminal
// check (via promptui's own readline dependency, already in go.mod — no new
// dependency added).
func IsInteractive(f *os.File) bool {
	return readline.IsTerminal(int(f.Fd()))
}
