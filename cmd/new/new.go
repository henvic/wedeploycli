package new

import (
	"context"
	"fmt"

	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmd/new/project"
	"github.com/wedeploy/cli/cmd/new/service"
	"github.com/wedeploy/cli/cmdflagsfromhost"
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
	Short:   "Create new project or install new service",
	PreRunE: preRun,
	RunE:    newRun,
	Args:    cobra.NoArgs,
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func newRun(cmd *cobra.Command, args []string) error {
	if setupHost.Service() != "" {
		return service.Cmd.RunE(cmd, []string{})
	}

	if setupHost.Project() != "" {
		return service.Cmd.RunE(cmd, []string{})
	}

	if !isterm.Check() {
		return cmd.Help()
	}

	fmt.Println(fancy.Question("Do you want to create a new project or install a new service?"))
	var options = fancy.Options{}
	options.Add("1", "project")
	options.Add("2", "service")

	switch option, err := options.Ask("What is your option"); option {
	case "1", "p", "project":
		return project.Cmd.RunE(cmd, []string{})
	case "2", "s", "service":
		return service.Cmd.RunE(cmd, []string{})
	default:
		return err
	}
}

func init() {
	setupHost.Init(NewCmd)

	NewCmd.AddCommand(project.Cmd)
	NewCmd.AddCommand(service.Cmd)
}