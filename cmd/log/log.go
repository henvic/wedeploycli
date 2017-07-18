package cmdlog

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
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
		Auth:    true,
		Project: true,
	},
	Pattern:               cmdflagsfromhost.FullHostPattern,
	UseProjectDirectory:   true,
	UseContainerDirectory: true,
}

func init() {
	setupHost.Init(LogCmd)
}

// LogCmd is used for getting logs about a given scope
var LogCmd = &cobra.Command{
	Use:     "log",
	Short:   "Show logs of the services",
	PreRunE: preRun,
	RunE:    logRun,
	Example: `  we log --project chat --container data
  we log --container data
  we log --project chat --container data
  we log --url data-chat.wedeploy.me
  we log --url data-chat.wedeploy.io --instance abc`,
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process()
}

func logRun(cmd *cobra.Command, args []string) error {
	var project = setupHost.Project()
	var container = setupHost.Container()

	level, levelErr := logs.GetLevel(severityArg)

	if levelErr != nil {
		return levelErr
	}

	if len(args) > 2 {
		return errors.New("invalid number of arguments")
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
		return "", errwrap.Wrapf("Can not parse since argument: {{err}}.", err)
	}

	// use microseconds instead of seconds (dashboard takes ms as a param)
	return fmt.Sprintf("%v000", since), err
}

func init() {
	LogCmd.Flags().StringVar(&instanceArg, "instance", "", `Instance (node) UID`)
	LogCmd.Flags().StringVar(&severityArg, "level", "0", `Severity (critical, error, warning, info (default), debug)`)
	LogCmd.Flags().StringVar(&sinceArg, "since", "", "Show since moment (i.e., 20min, 3h, UNIX timestamp)")
	LogCmd.Flags().BoolVarP(&watchArg, "watch", "w", false, "Watch / follow log output")
}
