package cmdlogs

import (
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/logs"
)

var (
	instanceArg string
	severityArg string
	sinceArg    string
	watchArg    bool
)

// LogsCmd is used for getting logs about a given scope
var LogsCmd = &cobra.Command{
	Use:   "logs [project] [container] --instance hash",
	Short: "Logs running on WeDeploy",
	Run:   logsRun,
	Example: `we logs (on project or container directory)
we logs chat
we logs portal email
we logs portal email --instance abc`,
}

func logsRun(cmd *cobra.Command, args []string) {
	c := cmdcontext.SplitArguments(args, 0, 2)

	project, container, err := cmdcontext.GetProjectOrContainerID(c)
	level, levelErr := logs.GetLevel(severityArg)

	if err != nil || len(args) > 2 || levelErr != nil {
		if err := cmd.Help(); err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	filter := &logs.Filter{
		Project:   project,
		Container: container,
		Instance:  instanceArg,
		Level:     level,
		Since:     getSince(),
	}

	switch watchArg {
	case true:
		logs.Watch(&logs.Watcher{
			Filter:          filter,
			PoolingInterval: time.Second,
		})
	default:
		logs.List(filter)
	}
}

func getSince() string {
	if sinceArg == "" {
		return ""
	}

	var since, sinceErr = logs.GetUnixTimestamp(sinceArg)

	if sinceErr != nil {
		sinceErr = errwrap.Wrapf("Can't parse since argument: {{err}}.", sinceErr)
		fmt.Fprintf(os.Stderr, "%v\n", sinceErr)
		os.Exit(1)
	}

	// use microseconds instead of seconds (dashboard takes ms as a param)
	return fmt.Sprintf("%v000", since)
}

func init() {
	LogsCmd.Flags().StringVar(&instanceArg, "instance", "", `Instance ID or hash`)
	LogsCmd.Flags().StringVar(&severityArg, "level", "0", `Severity (critical, error, warning, info (default), debug)`)
	LogsCmd.Flags().StringVar(&sinceArg, "since", "", "Show since moment (i.e., 20min, 3h, UNIX timestamp)")
	LogsCmd.Flags().BoolVarP(&watchArg, "watch", "w", false, "Watch / follow log output")
}
