// +build !windows

package run

import (
	"os"
	"syscall"
)

func runWait(container string) (*os.Process, error) {
	return os.StartProcess(getDockerPath(),
		[]string{bin, "wait", container},
		&os.ProcAttr{
			Sys: &syscall.SysProcAttr{
				Setpgid: true,
			},
			Files: []*os.File{nil, nil, nil},
		})
}
