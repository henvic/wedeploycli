package isterm

import (
	"os"

	"github.com/wedeploy/cli/envs"
	"golang.org/x/crypto/ssh/terminal"
)

// Check if user is using terminal
func Check() bool {
	if _, ok := os.LookupEnv(envs.SkipTerminalVerification); ok {
		return true
	}

	return terminal.IsTerminal(int(os.Stdin.Fd()))
}
