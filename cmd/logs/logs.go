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
	severityArg string
	sinceArg    string
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

	project, container, err := cmdcontext.GetProjectOrContainerID(c)
	level, levelErr := logs.GetLevel(severityArg)

	// 3rd argument might be instance ID
	if err != nil || len(args) > 3 || levelErr != nil {
		if err := cmd.Help(); err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	var logPath = []string{project, container}

	filter := &logs.Filter{
		Level: level,
		Since: getSince(),
	}

	switch followArg {
	case true:
		logs.Watch(&logs.Watcher{
			Filter:          filter,
			Paths:           logPath,
			PoolingInterval: time.Second,
		})
	default:
		logs.List(filter, logPath...)
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
	LogsCmd.Flags().StringVar(&severityArg, "level", "0", `Severity (critical, error, warning, info (default), debug)`)
	LogsCmd.Flags().StringVar(&sinceArg, "since", "", "Show since moment (i.e., 20min, 3h, UNIX timestamp)")
	LogsCmd.Flags().BoolVarP(&followArg, "follow", "f", false, "Follow log output")
}
