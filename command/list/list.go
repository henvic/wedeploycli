package list

import (
	"context"
	"fmt"

	"github.com/henvic/ctxsignal"
	"github.com/henvic/wedeploycli/cmdflagsfromhost"
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/command/internal/we"
	listinstances "github.com/henvic/wedeploycli/command/list/instances"
	listprojects "github.com/henvic/wedeploycli/command/list/projects"
	listservices "github.com/henvic/wedeploycli/command/list/services"
	"github.com/henvic/wedeploycli/list"
	"github.com/henvic/wedeploycli/projects"
	"github.com/henvic/wedeploycli/services"

	"github.com/spf13/cobra"
)

// ListCmd is used for getting a list of projects and services
var ListCmd = &cobra.Command{
	Use: "list",
	Example: `  lcp list --project chat --service data
  lcp list --url data-chat.lfr.cloud`,
	Short:   "Show list of projects and services",
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
		Auth: true,
	},
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

	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()

	l.Watch(ctx, we.Context())

	if _, err := ctxsignal.Closed(ctx); err == nil {
		fmt.Println()
	}

	return nil
}

func init() {
	setupHost.Init(ListCmd)

	ListCmd.Flags().BoolVarP(
		&detailed,
		"detailed", "d", false, "Show more services details")

	ListCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Show and watch for changes")
	ListCmd.AddCommand(listprojects.ListProjectsCmd)
	ListCmd.AddCommand(listservices.ListServicesCmd)
	ListCmd.AddCommand(listinstances.ListInstancesCmd)
}
