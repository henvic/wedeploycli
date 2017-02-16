package verbose

import (
	"fmt"
	"io"
	"os"

	"github.com/wedeploy/cli/color"
)

var (
	// Enabled flag
	Enabled = false

	// ErrStream is the stream for errors
	ErrStream io.Writer = os.Stderr

	unsafeVerbose = false
)

func init() {
	unsafeVerbose = IsUnsafeMode()
}

// IsUnsafeMode checks if the unsafe verbose mode is on
func IsUnsafeMode() bool {
	if unsafe, _ := os.LookupEnv("WEDEPLOY_UNSAFE_VERBOSE"); unsafe == "true" {
		return true
	}

	return false
}

// SafeEscape string
func SafeEscape(value string) string {
	if unsafeVerbose {
		return value
	}

	return color.Format(color.BgYellow, " hidden value ")
}

// SafeEscapeSlice of strings
func SafeEscapeSlice(values []string) string {
	if unsafeVerbose {
		return fmt.Sprintf("%v", values)
	}

	var plural string

	if len(values) != 1 {
		plural = "s"
	}

	return color.Format(color.BgYellow, " %d hidden value%s ", len(values), plural)
}

// Debug prints verbose messages to stderr on verbose mode
func Debug(a ...interface{}) {
	if Enabled {
		fmt.Fprintln(ErrStream, a...)
	}
}
