package cmdlist

import (
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/list"

	"github.com/spf13/cobra"
)

// ListCmd is used for getting a list of projects and containers
var ListCmd = &cobra.Command{
	Use: "list <host> or --project <project> --container <container>",
	Example: `we list --project chat --container data
we list --container data
we list --project chat --container data
we list --url data.chat.wedeploy.me
we list --url data.chat.wedeploy.io`,
	Short:   "List projects and containers running",
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
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process()
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
