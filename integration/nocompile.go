// +build nocompile

package integration

import (
	"fmt"
	"os"
)

func init() {
	_, _ = fmt.Fprintln(os.Stderr, `Skipping compilation: using "lcp" command available on system.`)
	binary = "lcp"
}
