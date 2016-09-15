// +build !windows

package run

import (
	"os/exec"
	"syscall"
)

func tryAddCommandToNewProcessGroup(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	cmd.SysProcAttr.Setpgid = true
	cmd.SysProcAttr.Pgid = 0
}
