package exec

import (
	"context"
	"strings"

	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/verbose"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/command/internal/we"
	"github.com/wedeploy/cli/isterm"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/shell"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern | cmdflagsfromhost.InstancePattern,

	Requires: cmdflagsfromhost.Requires{
		Project:  true,
		Service:  true,
		Instance: true,
	},

	AutoSelectSingleInstance: true,

	PromptMissingService:  true,
	PromptMissingInstance: true,
}

// ExecCmd executes a process (command) remotely
var ExecCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute command remotely",
	Example: `  lcp exec -p demo -s web -- ls
  lcp exec -p demo -s web --instance any -- uname -a (run command on any instance)
  lcp exec -p demo -s web --instance ab123 -- backup-db`,
	PreRunE: execPreRun,
	RunE:    execRun,
	Args:    cobra.MinimumNArgs(1),
	Hidden:  true,
}

func init() {
	setupHost.Init(ExecCmd)
}

func execPreRun(cmd *cobra.Command, args []string) error {
	if err := setupHost.Process(context.Background(), we.Context()); err != nil {
		return err
	}

	servicesClient := services.New(we.Context())

	_, err := servicesClient.Get(context.Background(), setupHost.Project(), setupHost.Service())
	return err
}

func execRun(cmd *cobra.Command, args []string) error {
	var childArgs []string

	if len(args) > 1 {
		childArgs = args[1:]
	}

	var wectx = we.Context()
	var host = wectx.Infrastructure()

	host = strings.Replace(host, "http://", "", 1)
	host = strings.Replace(host, "https://", "", 1)

	var params = shell.Params{
		Host:  host,
		Token: wectx.Token(),

		ProjectID: setupHost.Project(),
		ServiceID: setupHost.Service(),
		Instance:  setupHost.Instance(),

		AttachStdin: true,
		TTY:         isterm.Stdin(),
	}

	switch params.TTY {
	case true:
		verbose.Debug("Attaching tty")
	default:
		verbose.Debug("Not attaching tty")
	}

	return shell.Run(context.Background(), params, args[0], childArgs...)
}
