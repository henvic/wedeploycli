package shell

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/isterm"
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

// ShellCmd opens a shell remotely
var ShellCmd = &cobra.Command{
	Use:     "shell",
	Aliases: []string{"ssh"},
	Short:   "Opens a shell on a container of your service",
	PreRunE: shellPreRun,
	RunE:    shellRun,
	Args:    cobra.NoArgs,
}

func init() {
	setupHost.Init(ShellCmd)
}

func shellPreRun(cmd *cobra.Command, args []string) error {
	if !isterm.Check() {
		return errors.New("can't open terminal: tty wasn't found")
	}

	if err := setupHost.Process(context.Background(), we.Context()); err != nil {
		return err
	}

	wectx := we.Context()

	servicesClient := services.New(wectx)

	service, err := servicesClient.Get(context.Background(), setupHost.Project(), setupHost.Service())

	if err != nil {
		return err
	}

	fmt.Printf("You are accessing one instance of %s (total: %d)\n",
		color.Format(color.Bold, setupHost.Host()), service.Scale)

	fmt.Printf("%s\n\n", color.Format(color.FgYellow,
		"Warning: don't use this shell to make changes on your services. Only changes inside volumes persist."))

	return nil
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

		TTY: true,
	}

	return shell.Run(context.Background(), params, "")
}
