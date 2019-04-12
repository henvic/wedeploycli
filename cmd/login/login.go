package login

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/login"
	"github.com/wedeploy/cli/projects"
)

var noLaunchBrowser bool

// LoginCmd sets the user credential
var LoginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Login into your account",
	Args:    cobra.NoArgs,
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
	return setupHost.Process(context.Background(), we.Context())
}

func verifyAlreadyLoggedIn(wectx config.Context) error {
	projectsClient := projects.New(we.Context())
	_, err := projectsClient.List(context.Background())
	af, ok := err.(apihelper.APIFault)

	if ok && af.Status == http.StatusUnauthorized {
		_, _ = fmt.Fprintln(os.Stderr, fancy.Error(
			fmt.Sprintf(
				`Validating current token for %v on %v (%v) failed`,
				wectx.Username(),
				wectx.Remote(),
				wectx.InfrastructureDomain()),
		))
		return nil
	}

	return fmt.Errorf(`already logged in as %v on %v (%v)
Logout first with "liferay logout"`,
		wectx.Username(),
		wectx.Remote(),
		wectx.InfrastructureDomain())
}

func loginRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()

	if wectx.Username() != "" {
		if err := verifyAlreadyLoggedIn(wectx); err != nil {
			return err
		}
	}

	a := login.Authentication{
		NoLaunchBrowser: noLaunchBrowser,
		TipCommands:     true,
	}

	return a.Run(context.Background(), wectx)
}
