// +build !windows

package exechelper

import (
	"os/exec"
	"syscall"
)

// AddCommandToNewProcessGroup adds a command to a new process group
func AddCommandToNewProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	cmd.SysProcAttr.Setpgid = true
	cmd.SysProcAttr.Pgid = 0
}
