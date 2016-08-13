/*
Package errorhandling provides a error handling system to be used as
root.Execute() error handler. It should not be used somewhere else.
*/
package errorhandling

import (
	"fmt"
	"os"
	"runtime"
	"time"

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

// Handle error to a more friendly format
func Handle(err error) error {
	if err == nil {
		return nil
	}

	return (&handler{err}).handle()
}

type handler struct {
	err error
}

func (h *handler) handle() error {
	return h.err
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
