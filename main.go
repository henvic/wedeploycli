/*
api.cmd

	https://github.com/wedeploy/cli

*/

package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/wedeploy/cli/cmd"
	"github.com/wedeploy/cli/defaults"
)

const timeFormat = "Mon Jan _2 15:04:05 MST 2006"

const errorTemplate = `An unrecoverable error has occurred.
Please report this error at
https://github.com/wedeploy/cli/issues/

%s
Time: %s
%s`

func panickingListener(panicking *bool) {
	if !*panicking {
		return
	}

	var version = fmt.Sprintf("Version: %s %s/%s (runtime: %s)",
		defaults.Version,
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version())

	fmt.Fprintln(os.Stderr, color.RedString(errorTemplate,
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

func main() {
	var panicking = true
	defer panickingListener(&panicking)
	cmd.Execute()
	panicking = false
}
