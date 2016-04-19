package verbose

import (
	"fmt"
	"io"
	"os"
)

var (
	// Enabled flag
	Enabled = false

	// ErrStream is the stream for errors
	ErrStream io.Writer = os.Stderr
)

// Debug prints verbose messages to stderr on verbose mode
func Debug(a ...interface{}) {
	if Enabled {
		fmt.Fprintln(ErrStream, a...)
	}
}
