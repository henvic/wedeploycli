package isterm

import (
	"os"

	"github.com/wedeploy/cli/envs"
	"github.com/wedeploy/cli/verbose"
	"golang.org/x/crypto/ssh/terminal"
)

// NoTTY helps to simulate a non-terminal process
var NoTTY = false

// Stdin returns if stdin is connected to a terminal
func Stdin() bool {
	return check(os.Stdin)
}

// Stderr returns if stderr is connected to a terminal
func Stderr() bool {
	return check(os.Stderr)
}

// Stdout returns if stdout is connected to a terminal
func Stdout() bool {
	return check(os.Stdout)
}

// Check if stdin, stderr, and stdout are connected to a terminal
func Check() bool {
	return check(os.Stdin) && check(os.Stderr) && check(os.Stdout)
}

// Check if user is using terminal
func check(f *os.File) bool {
	if NoTTY {
		return false
	}

	_, skip := os.LookupEnv(envs.SkipTerminalVerification)
	is := terminal.IsTerminal(int(f.Fd()))

	if skip && !is {
		verbose.Debug("A terminal wasn't found, but system was told to ignore verification")
	}

	return skip || is
}
