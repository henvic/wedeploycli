// +build nocompile

package integration

import (
	"fmt"
	"os"
)

func init() {
	_, _ = fmt.Fprintln(os.Stderr, `Skipping compilation: using "we" command available on system.`)
	binary = "we"
}
