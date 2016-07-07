// +build windows

package run

import "os"

func runWait(container string) (*os.Process, error) {
	return os.StartProcess(getDockerPath(),
		[]string{bin, "wait", container},
		&os.ProcAttr{
			Files: []*os.File{nil, nil, nil},
		})
}
