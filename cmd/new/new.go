package new

import (
	"context"
	"fmt"

	"github.com/wedeploy/cli/cmd/canceled"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmd/new/project"
	"github.com/wedeploy/cli/cmd/new/service"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/fancy"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/isterm"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,
}

// NewCmd is used to create new projects/services
var NewCmd = &cobra.Command{
	Use:     "new",
	Short:   "Create new project or install new service\n\t\t",
	PreRunE: preRun,
	RunE:    newRun,
	Args:    cobra.NoArgs,
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func newRun(cmd *cobra.Command, args []string) error {
	var projectID = setupHost.Project()
	var serviceID = setupHost.Service()

	if serviceID != "" {
		return service.Run(projectID, serviceID, setupHost.ServiceDomain())
	}

	if projectID != "" {
		return project.Run(projectID)
	}

	if !isterm.Check() {
		return cmd.Help()
	}

	var options = fancy.Options{}
	options.Add("1", "Create a project")
	options.Add("2", "Install a service")
	options.Add("3", "Cancel")

	q := fmt.Sprintf("Do you want to %s a new project or install a new service?",
		color.Format(color.FgMagenta, color.Bold, "create"))

	switch option, err := options.Ask(q); option {
	case "1", "p", "project":
		return project.Cmd.RunE(cmd, []string{})
	case "2", "s", "service":
		return service.Cmd.RunE(cmd, []string{})
	case "3", "cancel":
		return canceled.Skip()
	default:
		return err
	}
}

func init() {
	setupHost.Init(NewCmd)

	NewCmd.AddCommand(project.Cmd)
	NewCmd.AddCommand(service.Cmd)
}
