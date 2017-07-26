package cmdlogin

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/login"
)

var noLaunchBrowser bool

// LoginCmd sets the user credential
var LoginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Login into your account",
	PreRunE: preRun,
	RunE:    loginRun,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.RemotePattern,
}

func init() {
	setupHost.Init(LoginCmd)
	LoginCmd.Flags().BoolVar(&noLaunchBrowser, "no-browser", false, "Perform the operation without opening your browser")
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process()
}

func loginRun(cmd *cobra.Command, args []string) error {
	if config.Context.Remote == defaults.LocalRemote {
		return errors.New("Local remote does not require logging in")
	}

	if config.Context.Username != "" {
		return fmt.Errorf(`Already logged in as %v on %v (%v)
Logout first with "we logout"`, config.Context.Username, config.Context.Remote, config.Context.InfrastructureDomain)
	}

	a := login.Authentication{
		NoLaunchBrowser: noLaunchBrowser,
	}

	return a.Run(context.Background())
}
