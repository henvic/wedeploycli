package cmdlist

import (
	"fmt"
	"os"

	"github.com/wedeploy/cli/list"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/projects"
)

// ListCmd is used for getting a list of projects and containers
var ListCmd = &cobra.Command{
	Use:   "list or list [project] to filter by project",
	Short: "List projects and containers running on WeDeploy",
	Run:   listRun,
}

var detailed bool

func listRun(cmd *cobra.Command, args []string) {
	var l = &list.List{
		Projects: []projects.Project{},
	}

	l.Detailed = detailed

	switch len(args) {
	case 0:
		var ps, err = projects.List()

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		for _, project := range ps {
			l.Projects = append(l.Projects, project)
		}
	case 1:
		var p, err = projects.Get(args[0])

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		l.Projects = append(l.Projects, p)
	default:
		println("This command takes 0 or 1 argument.")
		os.Exit(1)
	}

	l.Print()
}

func init() {
	ListCmd.Flags().BoolVarP(
		&detailed,
		"detailed", "d", false, "Show more containers details.")
}
