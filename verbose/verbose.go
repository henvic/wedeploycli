package verbose

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/envs"
)

var (
	// Enabled flag
	Enabled = false

	// Deferred flag to only print debugging on end of program execution
	Deferred = false

	// ErrStream is the stream for errors
	ErrStream io.Writer = os.Stderr

	unsafeVerbose = false

	bufDeferredVerbose      bytes.Buffer
	bufDeferredVerboseMutex sync.Mutex
)

func init() {
	unsafeVerbose = IsUnsafeMode()
}

// IsUnsafeMode checks if the unsafe verbose mode is on
func IsUnsafeMode() bool {
	if unsafe, _ := os.LookupEnv(envs.UnsafeVerbose); unsafe == "true" {
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

	if !Deferred {
		fmt.Fprintln(ErrStream, a...)
		return
	}

	bufDeferredVerboseMutex.Lock()
	bufDeferredVerbose.WriteString(fmt.Sprintln(a...))
	bufDeferredVerboseMutex.Unlock()
}

// PrintDeferred debug messages
func PrintDeferred() {
	if !Deferred {
		return
	}

	bufDeferredVerboseMutex.Lock()
	if bufDeferredVerbose.Len() != 0 {
		fmt.Fprintf(ErrStream, "\n%v\n", color.Format(color.BgHiBlue, " Deferred verbose messages below "))
		_, _ = bufDeferredVerbose.WriteTo(ErrStream)
	}
	bufDeferredVerboseMutex.Unlock()
}
