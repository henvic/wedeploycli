/*
Package errorhandling provides a error handling system to be used as
root.Execute() error handler or on watches. It should not be used somewhere else.
*/
package errorhandling

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/verbose"
)

const timeFormat = "Mon Jan _2 15:04:05 MST 2006"

const panicTemplate = `An unrecoverable error has occurred.
Please report this error at
https://github.com/wedeploy/cli/issues/

%s
Time: %s
%s`

// Handle error to a more friendly format
func Handle(cmdName string, err error) error {
	if err == nil {
		return nil
	}

	return (&handler{
		cmdName: cmdName,
		err:     err,
	}).handle()
}

type handler struct {
	cmdName string
	err     error
}

type reason struct {
	filter  reasonFilter
	reason  string
	message string
}

type reasonFilter struct {
	url      string
	urlRegex *regexp.Regexp
}

// this can't be a map because we don't want to rely on guaranteed order
// it stops at the first time a reliable message is found
type messages []reason

func init() {
	for k := range friendlyMessages {
		rm := friendlyMessages[k]

		for j, i := range rm {
			if i.filter.url != "" {
				friendlyMessages[k][j].filter.urlRegex = regexp.MustCompile(i.filter.url)
			}
		}
	}
}

func (m messages) tryGetMessage(cmd, uri string, reason string) (string, bool) {
	for _, fm := range m {
		if fm.reason != reason {
			continue
		}

		if fm.filter.urlRegex != nil {
			var u, err = url.Parse(uri)

			if err != nil {
				println("error handling failed to parse URL: " + err.Error())
				return "", false
			}

			var matches = fm.filter.urlRegex.FindStringSubmatch(u.Path)

			if len(matches) == 0 {
				return "", false
			}
		}

		if fm.message != "" {
			return fm.message, true
		}

		verbose.Debug("Missing friendly message for reason " + reason + " in command " + cmd)
	}

	return "", false
}

// tryGetMessage tries to get a human-friendly error message from the
// command / local (falling back to the global) error message lists.
// It uses an array to store the errors and stops
func tryGetMessage(cmd, uri string, reason string) (string, bool) {
	if _, ok := friendlyMessages[cmd]; ok {
		if msg, msgFound := friendlyMessages[cmd].tryGetMessage(cmd, uri, reason); msgFound {
			return msg, msgFound
		}
	}

	return friendlyMessages["we"].tryGetMessage(cmd, uri, reason)
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
		rtm, ok := tryGetMessage(h.cmdName, err.URL, e.Reason)
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
