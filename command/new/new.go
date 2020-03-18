package new

import (
	"context"
	"errors"
	"fmt"

	"github.com/henvic/wedeploycli/cmdflagsfromhost"
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/command/canceled"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/command/new/project"
	"github.com/henvic/wedeploycli/command/new/service"
	"github.com/henvic/wedeploycli/fancy"

	"github.com/henvic/wedeploycli/isterm"
	"github.com/spf13/cobra"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.RegionPattern | cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Auth: true,
	},
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
	var region = setupHost.Region()
	var projectID = setupHost.Project()
	var serviceID = setupHost.Service()

	if serviceID != "" && setupHost.Region() != "" {
		return errors.New("cannot use --region on this")
	}

	if serviceID != "" {
		return service.Run(projectID, serviceID, setupHost.ServiceDomain())
	}

	if projectID != "" {
		return project.Run(projectID, region)
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
		if setupHost.Region() != "" {
			return errors.New("cannot use --region on this")
		}

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
