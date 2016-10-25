package cmdlog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/logs"
)

var (
	instanceArg string
	severityArg string
	sinceArg    string
	watchArg    bool
)

var setupHost = cmdflagsfromhost.SetupHost{
	Requires: cmdflagsfromhost.Requires{
		Auth: true,
	},
	Pattern: cmdflagsfromhost.FullHostPattern,
	UseProjectDirectoryForContainer: true,
}

func init() {
	setupHost.Init(LogCmd)
}

// LogCmd is used for getting logs about a given scope
var LogCmd = &cobra.Command{
	Use:     "log <host> or --project <project> --container <container> --instance hash",
	Short:   "See logs of what is running on WeDeploy",
	PreRunE: preRun,
	RunE:    logRun,
	Example: `we log --project chat --container data
we log chat
we log data.chat
we log data.chat.wedeploy.me
we log data.chat.wedeploy.io --instance abc`,
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(args)
}

func logRun(cmd *cobra.Command, args []string) error {
	var project = setupHost.Project()
	var container = setupHost.Container()

	level, levelErr := logs.GetLevel(severityArg)

	if levelErr != nil {
		return levelErr
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
		if err = logs.List(context.Background(), filter); err != nil {
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
	LogCmd.Flags().StringVar(&instanceArg, "instance", "", `Instance ID or hash`)
	LogCmd.Flags().StringVar(&severityArg, "level", "0", `Severity (critical, error, warning, info (default), debug)`)
	LogCmd.Flags().StringVar(&sinceArg, "since", "", "Show since moment (i.e., 20min, 3h, UNIX timestamp)")
	LogCmd.Flags().BoolVarP(&watchArg, "watch", "w", false, "Watch / follow log output")
}
