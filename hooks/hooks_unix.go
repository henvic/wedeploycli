// +build !windows

package hooks

import "os/exec"

func run(command string) error {
	process := exec.Command("sh", "-c", command)
	process.Stderr = errStream
	process.Stdout = outStream
	return process.Run()
}
