package cmdlogs

import (
	"fmt"
	"os"
	"time"

	"github.com/launchpad-project/cli/cmdcontext"
	"github.com/launchpad-project/cli/logs"
	"github.com/spf13/cobra"
)

var (
	severityArg string
	sinceArg    int64
	followArg   bool
)

// LogsCmd is used for getting logs about a given scope
var LogsCmd = &cobra.Command{
	Use:   "logs [project] [container] [instance]",
	Short: "Logs running on WeDeploy",
	Run:   logsRun,
	Example: `we logs (on container directory)
we logs portal email
we logs portal email email5932`,
}

func logsRun(cmd *cobra.Command, args []string) {
	c := cmdcontext.SplitArguments(args, 0, 2)

	project, container, err := cmdcontext.GetProjectAndContainerID(c)
	level, levelErr := logs.GetLevel(severityArg)

	// 3rd argument might be instance ID
	if err != nil || len(args) > 3 || levelErr != nil {
		cmd.Help()
		os.Exit(1)
	}

	args[0] = project
	args[1] = container

	filter := &logs.Filter{
		Level: level,
		Since: fmt.Sprintf("%v", sinceArg),
	}

	switch followArg {
	case true:
		logs.Watch(&logs.Watcher{
			Filter:          filter,
			Paths:           args,
			PoolingInterval: time.Second,
		})
	default:
		logs.List(filter, args...)
	}
}

func init() {
	LogsCmd.Flags().StringVar(&severityArg, "level", "0", `Severity (critical, error, warning, info (default), debug)`)
	LogsCmd.Flags().Int64Var(&sinceArg, "since", 0, "Show logs since timestamp")
	LogsCmd.Flags().BoolVarP(&followArg, "follow", "f", false, "Follow log output")
}
