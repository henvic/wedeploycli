package verbose

import (
	"fmt"
	"io"
	"os"
)

var (
	// Enabled flag
	Enabled = false

	errStream io.Writer = os.Stderr
)

// Debug prints verbose messages to stderr on verbose mode
func Debug(a ...interface{}) {
	if Enabled {
		fmt.Fprintln(errStream, a...)
	}
}
