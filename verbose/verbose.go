package verbose

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/wedeploy/cli/color"
)

var (
	// Enabled flag
	Enabled = false

	// Defered flag to only print debugging on end of program execution
	Defered = false

	// ErrStream is the stream for errors
	ErrStream io.Writer = os.Stderr

	unsafeVerbose          = false
	bufDeferedVerbose      bytes.Buffer
	bufDeferedVerboseMutex sync.Mutex
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
	if !Enabled {
		return
	}

	if !Defered {
		fmt.Fprintln(ErrStream, a...)
		return
	}

	bufDeferedVerboseMutex.Lock()
	bufDeferedVerbose.WriteString(fmt.Sprintln(a...))
	bufDeferedVerboseMutex.Unlock()
}

// PrintDefered debug messages
func PrintDefered() {
	bufDeferedVerboseMutex.Lock()
	if bufDeferedVerbose.Len() != 0 {
		fmt.Fprintf(ErrStream, "\n%v\n", color.Format(color.BgHiBlue, " Defered verbose messages below "))
		bufDeferedVerbose.WriteTo(ErrStream)
	}
	bufDeferedVerboseMutex.Unlock()
}
