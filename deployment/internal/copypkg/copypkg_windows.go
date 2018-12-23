// +build windows

package copypkg

import (
	"context"
	"os"
	"os/exec"
)

// Copy file or directory.
func Copy(ctx context.Context, src, dest string) error {
	_, err := os.Stat(dest)

	if os.IsNotExist(err) {
		err = os.Mkdir(dest, 0700)
	}

	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, "robocopy", src, dest)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
