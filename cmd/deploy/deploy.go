package cmddeploy

import (
	"context"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/deploy/remote"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/defaults"
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
	Short:   "Perform services deployment",
	PreRunE: preRun,
	RunE:    runRun,
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	if err := checkNonRemoteSet(cmd); err != nil {
		return err
	}

	return setupHost.Process()
}

func checkNonRemoteSet(cmd *cobra.Command) error {
	var (
		u             = cmd.Flag("url").Value.String()
		remoteChanged = cmd.Flag("remote").Changed
		urlChanged    = cmd.Flag("url").Changed && strings.Contains(u, ".")
	)

	if !remoteChanged && !urlChanged {
		if err := cmd.Flag("remote").Value.Set(defaults.CloudRemote); err != nil {
			return errwrap.Wrapf("error setting default remote: {{err}}", err)
		}
		cmd.Flag("remote").Changed = true
	}

	return nil
}

func runRun(cmd *cobra.Command, args []string) error {
	var rd = &cmddeployremote.RemoteDeployment{
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
