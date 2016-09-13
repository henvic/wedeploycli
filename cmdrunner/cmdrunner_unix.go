// +build !windows

package cmdrunner

import (
	"os"
	"os/exec"
)

func run(command string) error {
	process := exec.Command("bash", "-c", command)
	process.Stderr = os.Stderr
	process.Stdout = os.Stdout
	return process.Run()
}
