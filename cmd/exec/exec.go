package exec

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/shell"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Project: true,
		Service: true,
	},

	PromptMissingService: true,
}

// ExecCmd executes a process (command) remotely
var ExecCmd = &cobra.Command{
	Use:     "exec",
	Short:   "Execute command remotely",
	PreRunE: execPreRun,
	RunE:    execRun,
	Args:    cobra.MinimumNArgs(1),
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

		TTY: true,
	}

	return shell.Run(context.Background(), params, args[0], childArgs...)
}