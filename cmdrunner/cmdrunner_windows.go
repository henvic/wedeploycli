// +build windows

package cmdrunner

import (
	"os"
	"os/exec"

	"github.com/wedeploy/cli/verbose"
)

// Run a process synchronously inheriting stderr and stdout
func run(command string) error {
	if checkBashExists() {
		return runBash(command)
	}

	verbose.Debug("Warning: bash not available, running command on Windows native cmd")
	return runCmd(command)
}

func runBash(command string) error {
	process := exec.Command("bash", "-c", command)
	process.Stderr = os.Stderr
	process.Stdout = os.Stdout
	return process.Run()
}

func runCmd(command string) error {
	process := exec.Command("cmd", "/c", command)
	process.Stderr = os.Stderr
	process.Stdout = os.Stdout
	return process.Run()
}

func checkBashExists() bool {
	_, err := exec.LookPath("bash")
	return err == nil
}
