package cmdactivities

import (
	"context"

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
	Requires: cmdflagsfromhost.Requires{
		Project: true,
		Auth:    true,
	},
	Pattern: cmdflagsfromhost.ProjectAndRemotePattern,
}

func init() {
	setupHost.Init(ActivitiesCmd)
	ActivitiesCmd.Flags().StringVar(&commit, "commit", "", "Filter by deployment hash")
	ActivitiesCmd.Flags().StringVar(&groupUID, "group", "", "Filter by Group UID")
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
