package cmdlist

import (
	"context"
	"time"

	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"

	"github.com/spf13/cobra"
)

// ListCmd is used for getting a list of projects and containers
var ListCmd = &cobra.Command{
	Use: "list --url <host>",
	Example: `  we list --project chat --container data
  we list --url data-chat.wedeploy.me
  we list --url data-chat.wedeploy.io`,
	Short:   "List deployments",
	PreRunE: preRun,
	Run:     listRun,
}

var (
	detailed bool
	watch    bool
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,
	UseProjectDirectoryForContainer: true,
	Requires: cmdflagsfromhost.Requires{
		Auth: true,
	},
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	if err := setupHost.Process(); err != nil {
		return err
	}

	if !watch {
		return checkProjectOrContainerExists()
	}

	return nil
}

func checkProjectOrContainerExists() (err error) {
	if setupHost.Container() != "" {
		if _, err = containers.Get(context.Background(), setupHost.Project(), setupHost.Container()); err != nil {
			return err
		}

		return nil
	}

	if setupHost.Project() != "" {
		_, err = projects.Get(context.Background(), setupHost.Project())
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

	if setupHost.Container() != "" {
		filter.Containers = []string{
			setupHost.Container(),
		}
	}

	var l = list.New(filter)

	l.Detailed = detailed

	if !watch {
		l.PoolingInterval = time.Minute
		l.StopCondition = alwaysStop
	}

	l.Start()
}

func init() {
	setupHost.Init(ListCmd)

	ListCmd.Flags().BoolVarP(
		&detailed,
		"detailed", "d", false, "Show more containers details")

	ListCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
}
