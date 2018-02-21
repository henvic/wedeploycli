package deploy

import (
	"context"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/deploy/remote"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
	},

	UseProjectFromWorkingDirectory: true,

	AllowMissingProject:        true,
	PromptMissingProject:       true,
	HideServicesPrompt:         true,
	AllowCreateProjectOnPrompt: true,
}

var quiet bool
var copyPackage string

// DeployCmd runs services
var DeployCmd = &cobra.Command{
	Use:     "deploy",
	Short:   "Deploy your services",
	Args:    cobra.NoArgs,
	PreRunE: preRun,
	RunE:    run,
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func run(cmd *cobra.Command, args []string) (err error) {
	if copyPackage != "" {
		if copyPackage, err = filepath.Abs(copyPackage); err != nil {
			return err
		}
	}

	var rd = &deployremote.RemoteDeployment{
		ProjectID: setupHost.Project(),
		ServiceID: setupHost.Service(),
		Remote:    setupHost.Remote(),

		CopyPackage: copyPackage,
		Quiet:       quiet,
	}

	_, err = rd.Run(context.Background())
	return err
}

func init() {
	DeployCmd.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"Deploy without watching status")
	DeployCmd.Flags().StringVar(&copyPackage, "copy-pkg", "",
		"Path to copy the deployment package to (for debugging)")
	_ = DeployCmd.Flags().MarkHidden("copy-pkg")

	setupHost.Init(DeployCmd)
}
