package shell

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/isterm"
	"github.com/wedeploy/cli/shell"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern | cmdflagsfromhost.InstancePattern,

	Requires: cmdflagsfromhost.Requires{
		Project:  true,
		Service:  true,
		Instance: true,
	},

	PromptMissingService:  true,
	PromptMissingInstance: true,
}

// ShellCmd opens a shell remotely
var ShellCmd = &cobra.Command{
	Use:     "shell",
	Aliases: []string{"ssh"},
	Short:   "Opens a shell on a container of your service\n\t\t",
	PreRunE: shellPreRun,
	RunE:    shellRun,
	Args:    cobra.NoArgs,
}

func init() {
	setupHost.Init(ShellCmd)
}

func shellPreRun(cmd *cobra.Command, args []string) error {
	if !isterm.Stdin() {
		return errors.New("can't open terminal: tty wasn't found")
	}

	return setupHost.Process(context.Background(), we.Context())
}

func shellRun(cmd *cobra.Command, args []string) error {
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
		TTY:         true,
	}

	return shell.Run(context.Background(), params, "")
}
