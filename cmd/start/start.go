package cmdstart

import (
	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/hooks"
)

// StartCmd starts the current project or container
var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Run start container hook",
	RunE:  startRun,
}

func init() {
	StartCmd.Hidden = true
}

func startRun(cmd *cobra.Command, args []string) error {
	var cp, err = containers.Read(".")

	if err != nil {
		return errwrap.Wrapf("container.json error: {{err}}", err)
	}

	if cp.Hooks == nil || (cp.Hooks.BeforeStart == "" &&
		cp.Hooks.Start == "" &&
		cp.Hooks.AfterStart == "") {
		println("> [" + cp.ID + "] has no start hooks")
		return nil
	}

	return cp.Hooks.Run(hooks.Start, ".", cp.ID)
}
