package util

import (
	"os"

	"golang.org/x/term"
)

// IsTTY reports whether stdout is an interactive terminal.
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// IsCI reports whether the process is running inside a CI environment.
func IsCI() bool {
	return os.Getenv("CI") == "true"
}
