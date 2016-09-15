// +build windows

package run

import "os/exec"

// Windows has no support for process groupss
// https://github.com/wedeploy/cli/issues/40
func tryAddCommandToNewProcessGroup(cmd *exec.Cmd) {}
