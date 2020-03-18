package login

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/henvic/wedeploycli/apihelper"
	"github.com/henvic/wedeploycli/cmdflagsfromhost"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/config"
	"github.com/henvic/wedeploycli/fancy"
	"github.com/henvic/wedeploycli/login"
	"github.com/henvic/wedeploycli/projects"
	"github.com/spf13/cobra"
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
Logout first with "lcp logout"`,
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
