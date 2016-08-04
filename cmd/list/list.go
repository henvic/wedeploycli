package cmdlist

import (
	"github.com/wedeploy/cli/list"

	"github.com/spf13/cobra"
)

// ListCmd is used for getting a list of projects and containers
var ListCmd = &cobra.Command{
	Use:   "list or list [project] to filter by project",
	Short: "List projects and containers running on WeDeploy",
	Run:   listRun,
}

var (
	detailed bool
	watch    bool
)

func listRun(cmd *cobra.Command, args []string) {
	var filter = list.Filter{}

	switch len(args) {
	case 0:
	case 1:
		filter.Project = args[0]
	default:
		filter.Project = args[0]
		filter.Containers = args[1:]
	}

	var l = list.New(filter)

	l.Detailed = detailed

	switch watch {
	case true:
		var w = list.NewWatcher(l)
		w.Start()
	default:
		l.Print()
	}
}

func init() {
	ListCmd.Flags().BoolVarP(
		&detailed,
		"detailed", "d", false, "Show more containers details.")

	ListCmd.Flags().BoolVar(
		&watch,
		"watch", false, "Watch for changes.")
}
