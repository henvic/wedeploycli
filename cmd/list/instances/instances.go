package listinstances

import (
	"context"
	"fmt"

	"github.com/henvic/ctxsignal"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/listinstances"
	"github.com/wedeploy/cli/services"
)

// ListInstancesCmd is used for getting a list of projects and services
var ListInstancesCmd = &cobra.Command{
	Use: "instances",
	Example: `  lcp list instances --project chat --service data
  lcp list instances --url data-chat.lfr.cloud`,
	Short:   "Show list of instances",
	Args:    cobra.NoArgs,
	PreRunE: preRun,
	RunE:    listRun,
}

var watch bool

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
		Service: true,
	},

	PromptMissingService: true,

	ListExtraDetails: list.Instances,
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func listRun(cmd *cobra.Command, args []string) error {
	servicesClient := services.New(we.Context())

	instances, err := servicesClient.Instances(context.Background(), setupHost.Project(), setupHost.Service())

	if err != nil {
		return err
	}

	if len(instances) == 0 {
		return fmt.Errorf("no instances found for service %s", setupHost.Service())
	}

	li := listinstances.New(setupHost.Project(), setupHost.Service())

	if !watch {
		return li.Once(context.Background(), we.Context())
	}

	fmt.Println(color.Format(color.FgHiBlack,
		"List of instances will be updated when a change occurs.\n"))

	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()

	li.Watch(ctx, we.Context())

	if _, err = ctxsignal.Closed(ctx); err == nil {
		fmt.Println()
	}

	return nil
}

func init() {
	setupHost.Init(ListInstancesCmd)
	ListInstancesCmd.Flags().BoolVarP(&watch, "watch", "w", false, "Show and watch for changes")
}
