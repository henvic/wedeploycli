package cmdlist

import (
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/list"

	"github.com/spf13/cobra"
)

// ListCmd is used for getting a list of projects and containers
var ListCmd = &cobra.Command{
	Use: "list <host> or --project <project> --container <container>",
	Example: `we list --project chat --container data
we list data
we list data.chat
we list data.chat.wedeploy.me
we list data.chat.wedeploy.io`,
	Short:   "List projects and containers running on WeDeploy",
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

func init() {
	setupHost.Init(ListCmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(args)
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

	if watch {
		list.NewWatcher(l).Start()
		return
	}

	l.Print()
}

func init() {
	ListCmd.Flags().BoolVarP(
		&detailed,
		"detailed", "d", false, "Show more containers details")

	ListCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Watch for changes")
}
