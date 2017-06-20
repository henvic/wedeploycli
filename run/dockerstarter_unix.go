// +build !windows,!darwin

package run

import (
	"bytes"
	"context"
	"os/exec"

	"github.com/wedeploy/cli/verbose"
)

func tryStartDocker() (err error) {
	var _, err = exec.LookPath("service")

	if err != nil {
		return err
	}

	var cmd = exec.CommandContext(context.Background(), "service", "docker", "start")
	var bufErr = &bytes.Buffer{}
	cmd.Stderr = bufErr
	err = cmd.Run()

	if err != nil {
		verbose.Debug(bufErr.String())
	}

	return err
}
