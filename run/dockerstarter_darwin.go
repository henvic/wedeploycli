// +build darwin

package run

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/wedeploy/cli/verbose"
)

func tryStartDocker() (err error) {
	var cmd = exec.CommandContext(context.Background(), "open", "-g", "-a", "Docker")
	var bufErr = &bytes.Buffer{}
	cmd.Stderr = bufErr
	err = cmd.Run()

	if err != nil {
		verbose.Debug(bufErr.String())
	}

	return err
}
