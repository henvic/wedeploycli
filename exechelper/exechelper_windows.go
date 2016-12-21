// +build windows

package exechelper

import "os/exec"

// AddCommandToNewProcessGroup is not used on Windows
// Windows has no support for process groups
// https://github.com/wedeploy/cli/issues/40
func AddCommandToNewProcessGroup(cmd *exec.Cmd) {}
