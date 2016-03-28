package cmdlogs

import (
	"fmt"
	"os"

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
	Short: "Logs running on Launchpad",
	Run:   logsRun,
	Example: `launchpad logs (on container directory)
launchpad logs portal email
launchpad logs portal email email5932`,
}

func logsRun(cmd *cobra.Command, args []string) {
	var instanceID string

	c := cmdcontext.SplitArguments(args, 0, 2)

	project, container, err := cmdcontext.GetProjectAndContainerID(c)

	level, levelErr := logs.GetLevel(severityArg)

	if err != nil || len(args) > 3 || levelErr != nil {
		cmd.Help()
		os.Exit(1)
	}

	if len(args) == 3 {
		instanceID = args[2]
	}

	args[0] = project
	args[1] = container

	filter := logs.Filter{
		Level:      level,
		InstanceID: instanceID,
		Since:      fmt.Sprintf("%v", sinceArg),
	}

	switch followArg {
	case true:
		logs.Watch(filter, args...)
	default:
		logs.List(filter, args...)
	}
}

func init() {
	LogsCmd.Flags().StringVar(&severityArg, "level", "0", `Severity (critical, error, warning, info (default), debug)`)
	LogsCmd.Flags().Int64Var(&sinceArg, "since", 0, "Show logs since timestamp")
	LogsCmd.Flags().BoolVarP(&followArg, "follow", "f", false, "Follow log output")
}
