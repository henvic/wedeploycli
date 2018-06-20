package listservices

import (
	"context"
	"fmt"

	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"

	"github.com/spf13/cobra"
)

// ListServicesCmd is used for getting a list of projects and services
var ListServicesCmd = &cobra.Command{
	Use: "services",
	Example: `  we list services --project chat --service data
   we list services --url data-chat.wedeploy.io`,
	Short:   "Show list of services",
	Args:    cobra.NoArgs,
	PreRunE: preRun,
	RunE:    listRun,
}

var (
	detailed bool
	watch    bool
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
	},

	HideServicesPrompt:   true,
	PromptMissingProject: true,
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := setupHost.Process(context.Background(), we.Context()); err != nil {
		return err
	}

	if !watch {
		return checkProjectOrServiceExists()
	}

	return nil
}

func checkProjectOrServiceExists() (err error) {
	if setupHost.Service() != "" {
		servicesClient := services.New(we.Context())
		// if service exists, project also exists
		_, err = servicesClient.Get(context.Background(), setupHost.Project(), setupHost.Service())
		return err
	}

	if setupHost.Project() != "" {
		projectsClient := projects.New(we.Context())
		_, err = projectsClient.Get(context.Background(), setupHost.Project())
		return err
	}

	return nil
}

func listRun(cmd *cobra.Command, args []string) error {
	var filter = list.Filter{
		Project: setupHost.Project(),
	}

	if setupHost.Service() != "" {
		filter.Services = []string{
			setupHost.Service(),
		}
	}

	var l = list.New(filter)

	if detailed {
		l.Details = list.Detailed
	}

	if !watch {
		return l.Once(context.Background(), we.Context())
	}

	fmt.Println(color.Format(color.FgHiBlack,
		"List of services will be updated when a change occurs.\n"))

	l.Start(context.Background(), we.Context())
	return nil
}

func init() {
	setupHost.Init(ListServicesCmd)

	ListServicesCmd.Flags().BoolVarP(
		&detailed,
		"detailed", "d", false, "Show more services details")

	ListServicesCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Show and watch for changes")
}
