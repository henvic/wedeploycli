/*
Package errorhandler provides a error handling system to be used as
root.Execute() error handler or on watches. It should not be used somewhere else.
*/
package errorhandler

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/templates"
	"github.com/wedeploy/cli/verbose"
)

const panicTemplate = `An unrecoverable error has occurred.
Please report this error at
https://github.com/wedeploy/cli/issues/

%s
Time: %s
%s`

// CommandName for the local message repository
var CommandName string

var afterError func()

// SetAfterError defines a function to run after global error
func SetAfterError(f func()) {
	afterError = f
}

// RunAfterError runs code after a global error, just before exiting
func RunAfterError() {
	if afterError != nil {
		afterError()
	}
}

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

func extractParentCommand(cmd string) string {
	var splitCmd = strings.Split(cmd, " ")
	return strings.Join(splitCmd[:len(splitCmd)-1], " ")
}

// tryGetPersonalizedMessage tries to get a human-friendly error message from the
// command / local error message lists falling back to the parent command
// and at last instance to the global
func tryGetPersonalizedMessage(cmd, reason string, context apihelper.APIFaultErrorContext) (string, bool) {
	local := cmd
getMessage:
	if haystack, ok := errorReasonCommandMessageOverrides[local]; ok {
		if msg, has := haystack[reason]; has {
			return msg, true
		}
	}

	if local = extractParentCommand(local); local != "" {
		goto getMessage
	}

	msg, ok := errorReasonMessage[reason]

	if !ok {
		return msg, ok
	}

	personalizedMsg, err := templates.Execute(msg, context)

	if err != nil {
		verbose.Debug(errwrap.Wrapf("error getting personalized message: {{err}}", err))
		return msg, ok
	}

	return personalizedMsg, ok
}

func (h *handler) unwrap() {
	// unwrapping has the side-effect of discarding wrapped messages
	// but we assume it is expected to do so (currently)
	var aerr = errwrap.GetType(h.err, new(apihelper.APIFault))

	if aerr != nil {
		h.err = aerr
	}
}

func (h *handler) handle() error {
	h.unwrap()

	if af := errwrap.GetType(h.err, apihelper.APIFault{}); af != nil {
		return h.handleAPIFaultError()
	}

	switch nerr, ok := h.err.(*url.Error); {
	case !ok:
		return h.err
	case nerr.Timeout():
		return errwrap.Wrapf("network connection timed out:\n{{err}}", h.err)
	default:
		return errwrap.Wrapf("network connection error:\n{{err}}", h.err)
	}
}

func (h *handler) handleAPIFaultError() error {
	var af = errwrap.GetType(h.err, apihelper.APIFault{})
	var err = af.(apihelper.APIFault)
	var msgs []string
	// we want to fallback to the default error if no friendly messages are found
	var anyFriendly bool

	for _, e := range err.Errors {
		rtm, ok := tryGetPersonalizedMessage(CommandName, e.Reason, e.Context)
		if ok {
			anyFriendly = true
			msgs = append(msgs, rtm)
		} else {
			msgs = append(msgs, e.Reason+": "+e.Context.Message())
		}

	}

	if !anyFriendly {
		return err
	}

	return errwrap.Wrap(errors.New(strings.Join(msgs, "\n")), err)
}

// GetTypes get a list of error types separated by ":"
func GetTypes(err error) string {
	var types []string

	errwrap.Walk(err, func(err error) {
		r := reflect.TypeOf(err)
		types = append(types, r.String())
	})

	return strings.Join(types, ":")
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

	_, _ = fmt.Fprintln(os.Stderr, color.Format(color.FgRed, panicTemplate,
		version,
		time.Now().Format(time.RubyDate), systemInfo()))
}

func systemInfo() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return fmt.Sprintf(`goroutines: %v | cgo calls: %v
CPUs: %v | Pointer lookups: %v
`, runtime.NumGoroutine(), runtime.NumCgoCall(), runtime.NumCPU(), m.Lookups)
}
