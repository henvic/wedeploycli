package isterm

import (
	"os"

	"github.com/wedeploy/cli/envs"
	"github.com/wedeploy/cli/verbose"
	"golang.org/x/crypto/ssh/terminal"
)

// NoTTY helps to simulate a non-terminal process
var NoTTY = false

// Check if user is using terminal
func Check() bool {
	if NoTTY {
		return false
	}

	_, skip := os.LookupEnv(envs.SkipTerminalVerification)
	is := terminal.IsTerminal(int(os.Stdin.Fd()))

	if skip && !is {
		verbose.Debug("A terminal wasn't found, but system was told to ignore verification")
	}

	return skip || is
}
