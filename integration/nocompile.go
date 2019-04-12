// +build nocompile

package integration

import (
	"fmt"
	"os"
)

func init() {
	_, _ = fmt.Fprintln(os.Stderr, `Skipping compilation: using "liferay" command available on system.`)
	binary = "liferay"
}
