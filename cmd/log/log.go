package log

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/ctxsignal"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/logs"
)

var (
	severity string
	since    string
	watch    bool
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern | cmdflagsfromhost.InstancePattern,

	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
	},

	PromptMissingProject: true,
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
	Args:    cobra.NoArgs,
	Example: `  lcp log --project chat --service data
  lcp log --service data
  lcp log --project chat --service data
  lcp log --url data-chat.lfr.cloud
  lcp log --url data-chat.lfr.cloud --instance 10ab22`,
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func logRun(cmd *cobra.Command, args []string) error {
	var project = setupHost.Project()
	var service = setupHost.Service()
	var instance = setupHost.Instance()

	var level, err = logs.GetLevel(severity)

	if err != nil {
		return err
	}

	if len(args) > 2 {
		return errors.New("invalid number of arguments")
	}

	var t string
	t, err = getSince()

	if err != nil {
		return err
	}

	f := &logs.Filter{
		Project:  project,
		Instance: instance,
		Level:    level,
		Since:    t,
	}

	if service != "" {
		f.Services = strings.Split(service, ",")
	}

	if !watch {
		logsClient := logs.New(we.Context())
		return logsClient.List(context.Background(), f)
	}

	watcher := &logs.Watcher{
		Filter:          f,
		PoolingInterval: time.Second,
	}

	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()

	watcher.Watch(ctx, we.Context())

	if _, err := ctxsignal.Closed(ctx); err == nil {
		fmt.Println()
	}

	return nil
}

func getSince() (string, error) {
	if since == "" {
		return "", nil
	}

	t, err := logs.GetUnixTimestamp(since)

	if err != nil {
		return "", errwrap.Wrapf("can't parse since argument: {{err}}.", err)
	}

	// use nanoseconds instead of seconds (console takes ns as a param)
	return fmt.Sprintf("%v000000000", t), err
}

func init() {
	LogCmd.Flags().StringVar(&severity, "level", "0", `Severity (critical, error, warning, info (default), debug)`)
	LogCmd.Flag("level").Hidden = true

	LogCmd.Flags().StringVar(&since, "since", "", "Show since moment (i.e., 20min, 3h, UNIX timestamp)")
	LogCmd.Flags().BoolVarP(&watch, "watch", "w", true, "Watch / follow log output")
	_ = LogCmd.Flags().MarkHidden("watch")
}
