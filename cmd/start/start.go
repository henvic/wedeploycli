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
	var container, err = containers.Read(".")

	if err != nil {
		return errwrap.Wrapf("container.json error: {{err}}", err)
	}

	if container.Hooks == nil || (container.Hooks.BeforeStart == "" &&
		container.Hooks.Start == "" &&
		container.Hooks.AfterStart == "") {
		println("> [" + container.ID + "] has no start hooks")
		return nil
	}

	return container.Hooks.Run(hooks.Start, ".", container.ID)
}
