package cmdlogs

import (
	"errors"
	"fmt"
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
	RunE:  logsRun,
	Example: `we logs (on project or container directory)
we logs chat
we logs portal email
we logs portal email --instance abc`,
}

func logsRun(cmd *cobra.Command, args []string) error {
	c := cmdcontext.SplitArguments(args, 0, 2)

	project, container, err := cmdcontext.GetProjectOrContainerID(c)

	if err != nil {
		return err
	}

	level, levelErr := logs.GetLevel(severityArg)

	if levelErr != nil {
		return err
	}

	if len(args) > 2 {
		return errors.New("Invalid number of arguments.")
	}

	since, err := getSince()

	if err != nil {
		return err
	}

	filter := &logs.Filter{
		Project:   project,
		Container: container,
		Instance:  instanceArg,
		Level:     level,
		Since:     since,
	}

	switch watchArg {
	case true:
		logs.Watch(&logs.Watcher{
			Filter:          filter,
			PoolingInterval: time.Second,
		})
	default:
		if err = logs.List(filter); err != nil {
			return err
		}
	}

	return nil
}

func getSince() (string, error) {
	if sinceArg == "" {
		return "", nil
	}

	var since, err = logs.GetUnixTimestamp(sinceArg)

	if err != nil {
		return "", errwrap.Wrapf("Can't parse since argument: {{err}}.", err)
	}

	// use microseconds instead of seconds (dashboard takes ms as a param)
	return fmt.Sprintf("%v000", since), err
}

func init() {
	LogsCmd.Flags().StringVar(&instanceArg, "instance", "", `Instance ID or hash`)
	LogsCmd.Flags().StringVar(&severityArg, "level", "0", `Severity (critical, error, warning, info (default), debug)`)
	LogsCmd.Flags().StringVar(&sinceArg, "since", "", "Show since moment (i.e., 20min, 3h, UNIX timestamp)")
	LogsCmd.Flags().BoolVarP(&watchArg, "watch", "w", false, "Watch / follow log output")
}
