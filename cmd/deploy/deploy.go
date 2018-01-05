package deploy

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/deploy/remote"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,
	Requires: cmdflagsfromhost.Requires{
		Auth: true,
	},
	AllowMissingProject: true,
}

var quiet bool

// DeployCmd runs services
var DeployCmd = &cobra.Command{
	Use:     "deploy",
	Short:   "Deploy your services",
	Args:    cobra.NoArgs,
	PreRunE: preRun,
	RunE:    runRun,
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(we.Context())
}

func runRun(cmd *cobra.Command, args []string) error {
	var rd = &deployremote.RemoteDeployment{
		ProjectID: setupHost.Project(),
		ServiceID: setupHost.Service(),
		Remote:    setupHost.Remote(),
		Quiet:     quiet,
	}

	var _, err = rd.Run(context.Background())
	return err
}

func init() {
	DeployCmd.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"Deploy without watching status")

	setupHost.Init(DeployCmd)
}
