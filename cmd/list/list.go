package list

import (
	"context"
	"fmt"
	"time"

	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"

	"github.com/spf13/cobra"
)

// ListCmd is used for getting a list of projects and services
var ListCmd = &cobra.Command{
	Use: "list",
	Example: `   we list --project chat --service data
   we list --url data-chat.wedeploy.io`,
	Short:   "Show list of projects and services",
	PreRunE: preRun,
	Run:     listRun,
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
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	if err := setupHost.Process(we.Context()); err != nil {
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

func alwaysStop() bool {
	return true
}

func listRun(cmd *cobra.Command, args []string) {
	var filter = list.Filter{
		Project: setupHost.Project(),
	}

	if setupHost.Service() != "" {
		filter.Services = []string{
			setupHost.Service(),
		}
	}

	var l = list.New(filter)

	l.Detailed = detailed

	if watch {
		fmt.Println(fancy.Info("--watch is in use. List of services will be updated when a change occurs."))
	}

	if !watch {
		l.PoolingInterval = time.Minute
		l.StopCondition = alwaysStop
	}

	l.Start(we.Context())
}

func init() {
	setupHost.Init(ListCmd)

	ListCmd.Flags().BoolVarP(
		&detailed,
		"detailed", "d", false, "Show more services details")

	ListCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Show and watch for changes")
}
