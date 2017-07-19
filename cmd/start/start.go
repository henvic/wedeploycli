package cmdstart

import (
	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/hooks"
)

// StartCmd starts the current project or service
var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Run start service hook",
	RunE:  startRun,
}

func init() {
	StartCmd.Hidden = true
}

func startRun(cmd *cobra.Command, args []string) error {
	var cp, err = services.Read(".")

	if err != nil {
		return errwrap.Wrapf("wedeploy.json error: {{err}}", err)
	}

	if cp.Hooks == nil || (cp.Hooks.BeforeStart == "" &&
		cp.Hooks.Start == "" &&
		cp.Hooks.AfterStart == "") {
		println("> [" + cp.ID + "] has no start hooks")
		return nil
	}

	return cp.Hooks.Run(hooks.Start, ".", cp.ID)
}
