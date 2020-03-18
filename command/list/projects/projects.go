package listprojects

import (
	"context"
	"fmt"

	"github.com/henvic/wedeploycli/cmdflagsfromhost"
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/list"

	"github.com/spf13/cobra"
)

// ListProjectsCmd is used for getting a list of projects
var ListProjectsCmd = &cobra.Command{
	Use:     "projects",
	Example: `  lcp list projects --url lfr.cloud`,
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

	if detailed {
		l.Details = list.Detailed
	}

	if !watch {
		return l.Once(context.Background(), we.Context())
	}

	fmt.Println(color.Format(color.FgHiBlack,
		"List of projects will be updated when a change occurs.\n"))

	l.Watch(context.Background(), we.Context())
	return nil
}

func init() {
	setupHost.Init(ListProjectsCmd)

	ListProjectsCmd.Flags().BoolVarP(
		&detailed,
		"detailed", "d", false, "Show more projects details")

	ListProjectsCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Show and watch for changes")
}
