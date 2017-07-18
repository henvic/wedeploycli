package cmddeploy

import (
	"fmt"
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
	UseProjectDirectoryForContainer: true,
}

var (
	quiet bool
)

// DeployCmd runs the WeDeploy local infrastructure
// and / or a project or container
var DeployCmd = &cobra.Command{
	Use:     "deploy",
	Short:   "Perform services deployment",
	PreRunE: preRun,
	RunE:    runRun,
}

func preRun(cmd *cobra.Command, args []string) error {
	if stopLocalInfraTmp {
		infra = false
	}

	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	if err := checkNonRemoteSet(cmd); err != nil {
		return err
	}

	if err := setupHost.Process(); err != nil {
		return err
	}

	return nil
}

func checkNonRemoteSet(cmd *cobra.Command) error {
	var u = cmd.Flag("url").Value.String()
	if u == "wedeploy.me" ||
		strings.Contains(u, ".wedeploy.me") ||
		cmd.Flag("remote").Value.String() == defaults.LocalRemote {
		return nil
	}

	var (
		remoteChanged = cmd.Flag("remote").Changed
		urlChanged    = cmd.Flag("url").Changed && strings.Contains(u, ".")
	)

	if remoteChanged || urlChanged {
		return checkNoLocalOnlyFlags(cmd)
	}

	if !remoteChanged && !urlChanged && hasActionLocalFlag() {
		if err := cmd.Flag("remote").Value.Set(defaults.LocalRemote); err != nil {
			return errwrap.Wrapf("error setting remote to local: {{err}}", err)
		}
		cmd.Flag("remote").Changed = true
	}

	return nil
}

func hasActionLocalFlag() bool {
	return isCommand("--dry-run-local-infra") ||
		isCommand("--start-local-infra") ||
		isCommand("--stop-local-infra")
}

func checkNoLocalOnlyFlags(cmd *cobra.Command) error {
	var localOnlyFlags = []string{
		"debug",
		"dry-run-local-infra",
		"start-local-infra",
		"stop-local-infra",
		"skip-local-infra",
		"experimental-image",
	}

	for _, k := range localOnlyFlags {
		f := cmd.Flag(k)
		if f.Changed {
			return fmt.Errorf("Flag --%v can only be used with the local remote (using: %v)",
				k,
				setupHost.Remote())
		}
	}

	return nil
}

func runRun(cmd *cobra.Command, args []string) error {
	if setupHost.Remote() == defaults.LocalRemote {
		return runLocal(cmd)
	}

	var rd = &cmddeployremote.RemoteDeployment{
		ProjectID: setupHost.Project(),
		ServiceID: setupHost.Container(),
		Remote:    setupHost.Remote(),
		Quiet:     quiet,
	}

	var _, err = rd.Run()
	return err
}

func init() {
	DeployCmd.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"Deploy without watching status")

	setupHost.Init(DeployCmd)
}
