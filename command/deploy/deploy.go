package deploy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/ctxsignal"
	"github.com/henvic/wedeploycli/cmdflagsfromhost"
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/command/deploy/internal/getproject"
	deployremote "github.com/henvic/wedeploycli/command/deploy/remote"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/deployment"
	"github.com/henvic/wedeploycli/jsonerror"
	"github.com/henvic/wedeploycli/logs"
	"github.com/henvic/wedeploycli/services"
	"github.com/spf13/cobra"
)

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.RegionPattern | cmdflagsfromhost.FullHostPattern,

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
	params       deployment.Params
	metadata     string
	follow       bool
	experimental bool
)

// DeployCmd runs services
var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your services",
	Example: `  lcp deploy
  lcp deploy https://gitlab.com/user/repo
  lcp deploy user/repo#branch`,
	Args:    cobra.MaximumNArgs(1),
	PreRunE: preRun,
	RunE:    run,
}

func checkMetadata() error {
	if metadata == "" {
		return nil
	}

	if err := json.Unmarshal([]byte(metadata), &deployment.Metadata{}); err != nil {
		return errwrap.Wrapf(
			"error parsing metadata: {{err}}",
			jsonerror.FriendlyUnmarshal(err))
	}

	return nil
}

func preRun(cmd *cobra.Command, args []string) error {
	params.Quiet = params.Quiet || params.SkipProgress // be quieter on skip progress as well
	params.Metadata = json.RawMessage(metadata)

	if err := checkMetadata(); err != nil {
		return err
	}

	if err := maybePreRunDeployFromGitRepo(cmd, args); err != nil {
		return err
	}

	return setupHost.Process(context.Background(), we.Context())
}

func run(cmd *cobra.Command, args []string) (err error) {
	params.Remote = setupHost.Remote()

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

func handleCopyPackage() (err error) {
	if params.CopyPackage == "" {
		return nil
	}

	if _, err = os.Stat(params.CopyPackage); err != nil {
		return errwrap.Wrapf("invalid --copy-pkg value: {{err}}", err)
	}

	params.CopyPackage, err = filepath.Abs(params.CopyPackage)
	return err
}

func local() (sil services.ServiceInfoList, err error) {
	if err = handleCopyPackage(); err != nil {
		return sil, err
	}

	params.ProjectID = setupHost.Project()
	params.Region = setupHost.Region()
	params.ServiceID = setupHost.Service()

	var rd = &deployremote.RemoteDeployment{
		Params:       params,
		Experimental: experimental,
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

	if params.CopyPackage != "" {
		return errors.New("can't create a local package when deploying with a git remote")
	}

	return nil
}

func fromGitRepo(repo string) (services.ServiceInfoList, error) {
	if params.Image != "" {
		return nil, errors.New("overwriting image when deploying from a git repository is not supported")
	}

	if metadata != "" {
		return nil, errors.New("using metadata when deploying from a git repository is not supported")
	}

	var err error
	params.Region = setupHost.Region()
	params.ProjectID, err = getproject.MaybeID(setupHost.Project(), params.Region)

	if err != nil {
		return nil, err
	}

	return deployment.DeployFromGitRepository(context.Background(), we.Context(), params, repo)
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
		Filter: f,
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
	DeployCmd.Flags().StringVar(&params.Image, "image", "", "Use different image for service")
	DeployCmd.Flags().StringVar(&metadata, "metadata", "", "Metadata in JSON")
	DeployCmd.Flags().BoolVar(&params.OnlyBuild, "only-build", false,
		"Skip deployment (only build)")
	DeployCmd.Flags().BoolVar(&params.SkipProgress, "skip-progress", false,
		"Skip watching deployment progress, quiet")
	DeployCmd.Flags().BoolVarP(&params.Quiet, "quiet", "q", false,
		"Suppress progress animations")
	DeployCmd.Flags().BoolVar(&follow, "follow", false,
		"Follow logs after deployment")
	DeployCmd.Flags().BoolVar(
		&experimental,
		"experimental", false, "Enable experimental deployment")
	DeployCmd.Flags().StringVar(&params.CopyPackage, "copy-pkg", "",
		"Path to copy the deployment package to (for debugging)")
	_ = DeployCmd.Flags().MarkHidden("metadata")
	_ = DeployCmd.Flags().MarkHidden("follow")
	_ = DeployCmd.Flags().MarkHidden("experimental")
	_ = DeployCmd.Flags().MarkHidden("copy-pkg")

	setupHost.Init(DeployCmd)
}
