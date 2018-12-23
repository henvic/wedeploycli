package deploy

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/henvic/ctxsignal"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/deploy/internal/getproject"
	"github.com/wedeploy/cli/cmd/deploy/remote"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/deployment"
	"github.com/wedeploy/cli/logs"
	"github.com/wedeploy/cli/services"
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

var (
	image        string
	onlyBuild    bool
	skipProgress bool
	quiet        bool
	follow       bool
	experimental bool
	copyPackage  string
)

// DeployCmd runs services
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your services",
	Example: `  we deploy
  we deploy https://gitlab.com/user/repo
  we deploy user/repo#branch`,
	Args:    cobra.MaximumNArgs(1),
	PreRunE: preRun,
	RunE:    run,
}

func preRun(cmd *cobra.Command, args []string) error {
	quiet = quiet || skipProgress // --quiet on skip progress; it also leads to a quieter output

	if err := maybePreRunDeployFromGitRepo(cmd, args); err != nil {
		return err
	}

	return setupHost.Process(context.Background(), we.Context())
}

func run(cmd *cobra.Command, args []string) (err error) {
	var sil services.ServiceInfoList
	switch {
	case len(args) != 0:
		sil, err = fromGitRepo(args[0])
	default:
		sil, err = local()
	}

	if err != nil || !follow {
		return err
	}

	return followLogs(sil)
}

func local() (sil services.ServiceInfoList, err error) {
	if copyPackage != "" {
		if copyPackage, err = filepath.Abs(copyPackage); err != nil {
			return nil, err
		}
	}

	var rd = &deployremote.RemoteDeployment{
		ProjectID: setupHost.Project(),
		ServiceID: setupHost.Service(),
		Remote:    setupHost.Remote(),

		Image: image,

		Experimental: experimental,
		CopyPackage:  copyPackage,

		OnlyBuild:    onlyBuild,
		SkipProgress: skipProgress,
		Quiet:        quiet,
	}

	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()

	var f deployremote.Feedback
	f, err = rd.Run(ctx)
	return f.Services, err
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

func fromGitRepo(repo string) (services.ServiceInfoList, error) {
	if image != "" {
		return nil, errors.New("overwriting image when deploying from a git repository is not supported")
	}

	projectID, err := getproject.MaybeID(setupHost.Project())

	if err != nil {
		return nil, err
	}

	params := deployment.ParamsFromRepository{
		ProjectID:  projectID,
		Repository: repo,

		OnlyBuild:    onlyBuild,
		SkipProgress: skipProgress,
		Quiet:        quiet,
	}

	return deployment.DeployFromGitRepository(context.Background(), we.Context(), params)
}

func followLogs(sil services.ServiceInfoList) error {
	now := time.Now()
	since := 10 * time.Second
	// BUG(henvic): this preliminary version only loads logs from 10s ago onwards.

	if len(sil) == 0 {
		panic("no services found to list logs")
	}

	var projectID = sil[0].ProjectID

	f := &logs.Filter{
		Project:  projectID,
		Services: sil.GetIDs(),

		Since: fmt.Sprintf("%v000000000", now.Add(-since).Unix()),
	}

	watcher := &logs.Watcher{
		Filter:          f,
		PoolingInterval: time.Second,
	}

	ctx, cancel := ctxsignal.WithTermination(context.Background())
	defer cancel()

	fmt.Println()
	fmt.Println(color.Format(color.FgBlue, color.Bold,
		fmt.Sprintf("Showing logs from %v ago onwards.", since)))
	fmt.Printf("You can exit anytime.\n\n")
	time.Sleep(220 * time.Millisecond)

	watcher.Watch(ctx, we.Context())

	if _, err := ctxsignal.Closed(ctx); err == nil {
		fmt.Println()
	}

	return nil
}

func init() {
	DeployCmd.Flags().StringVar(&image, "image", "", "Use different image for service")
	DeployCmd.Flags().BoolVar(&onlyBuild, "only-build", false,
		"Skip deployment (only build)")
	DeployCmd.Flags().BoolVar(&skipProgress, "skip-progress", false,
		"Skip watching deployment progress, quiet")
	DeployCmd.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"Suppress progress animations")
	DeployCmd.Flags().BoolVar(&follow, "follow", false,
		"Follow logs after deployment")
	DeployCmd.Flags().BoolVar(
		&experimental,
		"experimental", false, "Enable experimental deployment")
	DeployCmd.Flags().StringVar(&copyPackage, "copy-pkg", "",
		"Path to copy the deployment package to (for debugging)")
	_ = DeployCmd.Flags().MarkHidden("follow")
	_ = DeployCmd.Flags().MarkHidden("experimental")
	_ = DeployCmd.Flags().MarkHidden("copy-pkg")

	setupHost.Init(DeployCmd)
}
