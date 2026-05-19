package tty

import (
	"os"

	"golang.org/x/term"
)

// IsTerminal reports whether stdout is connected to a terminal.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
