/*
Package errorhandling provides a error handling system to be used as
root.Execute() error handler or on watches. It should not be used somewhere else.
*/
package errorhandling

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/defaults"
)

const timeFormat = "Mon Jan _2 15:04:05 MST 2006"

const panicTemplate = `An unrecoverable error has occurred.
Please report this error at
https://github.com/wedeploy/cli/issues/

%s
Time: %s
%s`

var CommandName string

// Handle error to a more friendly format
func Handle(err error) error {
	if err == nil {
		return nil
	}

	return (&handler{
		err: err,
	}).handle()
}

type handler struct {
	err error
}

type messages map[string]string

// tryGetMessage tries to get a human-friendly error message from the
// command / local (falling back to the global) error message lists.
func tryGetMessage(cmd, reason string) (string, bool) {
	if haystack, ok := errorReasonCommandMessageOverrides[cmd]; ok {
		if msg, has := haystack[reason]; has {
			return msg, true
		}
	}

	msg, ok := errorReasonMessage[reason]
	return msg, ok
}

func (h *handler) handle() error {
	switch h.err.(type) {
	case *apihelper.APIFault:
		return h.handleAPIFaultError()
	default:
		return h.err
	}
}

func (h *handler) handleAPIFaultError() error {
	var err = h.err.(*apihelper.APIFault)
	var msgs []string
	// we want to fallback to the default error if no friendly messages are found
	var anyFriendly bool

	for _, e := range err.Errors {
		rtm, ok := tryGetMessage(CommandName, e.Reason)
		if ok {
			anyFriendly = true
			msgs = append(msgs, rtm)
		} else {
			msgs = append(msgs, e.Reason+": "+e.Message)
		}

	}

	if !anyFriendly {
		return err
	}

	return errors.New(strings.Join(msgs, "\n"))
}

// Info prints useful system information for debugging
func Info() {
	var version = fmt.Sprintf("Version: %s %s/%s (runtime: %s)",
		defaults.Version,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version())

	if defaults.Build != "" {
		version += "\nbuild:" + defaults.Build
	}

	fmt.Fprintln(os.Stderr, color.Format(color.FgRed, panicTemplate,
		version,
		time.Now().Format(timeFormat), systemInfo()))
}

func systemInfo() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return fmt.Sprintf(`goroutines: %v | cgo calls: %v
CPUs: %v | Pointer lookups: %v
`, runtime.NumGoroutine(), runtime.NumCgoCall(), runtime.NumCPU(), m.Lookups)
}
