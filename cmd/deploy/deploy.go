package deploy

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/deploy/internal/getproject"
	"github.com/wedeploy/cli/cmd/deploy/remote"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/deployment"
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
	Use:   "deploy",
	Short: "Deploy your services",
	Example: `  we deploy
  we deploy <git repository address>
  we deploy wedeploy-examples/wordpress-example (on GitHub)
  we deploy https://github.com/wedeploy-examples/supermarket-web-example`,
	Args:    cobra.MaximumNArgs(1),
	PreRunE: preRun,
	RunE:    run,
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := maybePreRunDeployFromGitRepo(cmd, args); err != nil {
		return err
	}

	return setupHost.Process(context.Background(), we.Context())
}

func run(cmd *cobra.Command, args []string) (err error) {
	if len(args) != 0 {
		return deployFromGitRepo(args[0])
	}

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

func maybePreRunDeployFromGitRepo(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return nil
	}

	if s, _ := cmd.Flags().GetString("service"); s != "" {
		return errors.New("deploying with custom service ids isn't supported using git repositories")
	}

	if copyPackage != "" {
		return errors.New("can't create a local package when deploying with a git remote")
	}

	return nil
}

func deployFromGitRepo(repo string) error {
	projectID, err := getproject.MaybeID(setupHost.Project())

	if err != nil {
		return err
	}

	params := deployment.ParamsFromRepository{
		ProjectID:  projectID,
		Repository: repo,
		Quiet:      quiet,
	}

	return deployment.DeployFromGitRepository(context.Background(), we.Context(), params)
}

func init() {
	DeployCmd.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"Deploy without watching status")
	DeployCmd.Flags().StringVar(&copyPackage, "copy-pkg", "",
		"Path to copy the deployment package to (for debugging)")
	_ = DeployCmd.Flags().MarkHidden("copy-pkg")

	setupHost.Init(DeployCmd)
}
