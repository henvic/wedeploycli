// +build windows

package hooks

import (
	"os/exec"

	"github.com/wedeploy/cli/verbose"
)

// Run a process synchronously inheriting stderr and stdout
func run(command string) error {
	if checkShExists() {
		return runSh(command)
	}

	if checkBashExists() {
		return runBash(command)
	}

	verbose.Debug("Warning: bash not available, running hook on Windows native cmd")
	return runCmd(command)
}

func runSh(command string) error {
	process := exec.Command("sh", "-c", command)
	process.Stderr = errStream
	process.Stdout = outStream
	return process.Run()
}

func runBash(command string) error {
	process := exec.Command("bash", "-c", command)
	process.Stderr = errStream
	process.Stdout = outStream
	return process.Run()
}

func runCmd(command string) error {
	process := exec.Command("cmd", "/c", command)
	process.Stderr = errStream
	process.Stdout = outStream
	return process.Run()
}

func checkBashExists() bool {
	_, err := exec.LookPath("bash")
	return err == nil
}
