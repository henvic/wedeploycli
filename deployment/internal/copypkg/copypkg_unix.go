// +build !windows

package copypkg

import (
	"context"
	"os"
	"os/exec"
)

// Copy file or directory.
func Copy(ctx context.Context, src, dest string) error {
	cmd := exec.CommandContext(ctx, "cp", "-r", src, dest) // #nosec
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
