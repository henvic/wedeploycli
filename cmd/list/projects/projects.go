package listprojects

import (
	"context"
	"fmt"

	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/list"

	"github.com/spf13/cobra"
)

// ListProjectsCmd is used for getting a list of projects
var ListProjectsCmd = &cobra.Command{
	Use:     "projects",
	Example: `  we list projects --url wedeploy.io`,
	Short:   "Show list of projects",
	Args:    cobra.NoArgs,
	PreRunE: preRun,
	RunE:    listRun,
}

var (
	detailed bool
	watch    bool
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.RemotePattern,

	Requires: cmdflagsfromhost.Requires{
		Auth: true,
	},
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func listRun(cmd *cobra.Command, args []string) error {
	var filter = list.Filter{
		Project:      setupHost.Project(),
		HideServices: true,
	}

	var l = list.New(filter)

	l.Detailed = detailed

	if !watch {
		return l.Once(context.Background(), we.Context())
	}

	fmt.Println(color.Format(color.FgHiBlack,
		"List of projects will be updated when a change occurs.\n"))

	l.Start(context.Background(), we.Context())
	return nil
}

func init() {
	setupHost.Init(ListProjectsCmd)

	ListProjectsCmd.Flags().BoolVarP(
		&detailed,
		"detailed", "d", false, "Show more projects details")

	ListProjectsCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Show and watch for changes")
}