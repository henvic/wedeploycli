package cmdactivities

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/activities"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
)

// ActivitiesCmd runs the WeDeploy structure for development locally
var ActivitiesCmd = &cobra.Command{
	Use:     "activities",
	Short:   "List activities of a recent deployment",
	PreRunE: preRun,
	RunE:    activitiesRun,
	Hidden:  true,
}

var watchDeployCmd = &cobra.Command{
	Use:     "watch-deploy",
	Short:   "Watch deployment status",
	PreRunE: preRun,
	RunE:    watchDeployRun,
}

var (
	commit   string
	groupUID string
	services string
)

func preRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	if err := setupHost.Process(); err != nil {
		return err
	}

	return nil
}

var setupHost = cmdflagsfromhost.SetupHost{
	UseProjectDirectory: true,
	Requires: cmdflagsfromhost.Requires{
		Project: true,
		Auth:    true,
	},
	Pattern: cmdflagsfromhost.ProjectAndRemotePattern,
}

func init() {
	setupHost.Init(ActivitiesCmd)
	setupHost.Init(watchDeployCmd)
	ActivitiesCmd.AddCommand(watchDeployCmd)
	ActivitiesCmd.Flags().StringVar(&commit, "commit", "", "Filter by deployment hash")
	ActivitiesCmd.Flags().StringVar(&groupUID, "group", "", "Filter by Group UID")
	watchDeployCmd.Flags().StringVar(&commit, "commit", "", "Filter by deployment hash")
	watchDeployCmd.Flags().StringVar(&groupUID, "group", "", "Filter by Group UID")
	watchDeployCmd.Flags().StringVar(&services, "services", "", "Comma-separated services to watch")
}

func activitiesRun(cmd *cobra.Command, args []string) (err error) {
	var as []activities.Activity
	var f = activities.Filter{
		Commit:   commit,
		GroupUID: groupUID,
	}

	as, err = activities.List(context.Background(), setupHost.Project(), f)

	if err != nil {
		return err
	}

	activities.PrettyPrintList(as)
	return nil
}

func watchDeployRun(cmd *cobra.Command, args []string) (err error) {
	var servicesSlice = []string{}

	if services != "" {
		servicesSlice = strings.Split(services, ",")
	}

	var watcher = activities.NewDeployWatcher(
		context.Background(),
		setupHost.Project(),
		servicesSlice,
		activities.Filter{
			GroupUID: groupUID,
		},
	)

	return watcher.Run()
}
