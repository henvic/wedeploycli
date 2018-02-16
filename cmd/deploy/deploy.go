package deploy

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/deploy/remote"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/inspector"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
	},

	AllowMissingProject:        true,
	PromptMissingProject:       true,
	HideServicesPrompt:         true,
	AllowCreateProjectOnPrompt: true,
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

func validateWedeployJSONs() error {
	wd, err := os.Getwd()

	if err != nil {
		return err
	}

	_, err = inspector.InspectContext("", wd)
	return err
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := validateWedeployJSONs(); err != nil {
		return err
	}

	return setupHost.Process(context.Background(), we.Context())
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
