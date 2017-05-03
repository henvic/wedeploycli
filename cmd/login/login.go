package cmdlogin

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/loginserver"
	"github.com/wedeploy/cli/prompt"
	"github.com/wedeploy/cli/verbose"
)

var noLaunchBrowser bool

// LoginCmd sets the user credential
var LoginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Authenticate on WeDeploy",
	PreRunE: preRun,
	RunE:    loginRun,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.RemotePattern,
}

func init() {
	setupHost.Init(LoginCmd)
	LoginCmd.Flags().BoolVar(&noLaunchBrowser, "no-launch-browser", false, "Do not launch browser for authentication")
}

func maybeOpenBrowser(loginURL string) {
	fmt.Fprintf(os.Stdout, `Your browser is being open to visit:
	%v

`,
		loginURL)

	time.Sleep(420 * time.Millisecond)
	fmt.Println("Waiting authentication via browser")
	time.Sleep(710 * time.Millisecond)

	if err := browser.OpenURL(loginURL); err != nil {
		verbose.Debug("Can not open browser: " + err.Error())
		fmt.Println("Authenticate the CLI tool by visiting the link above.")
	}
}

func basicAuthLogin() error {
	var (
		username string
		password string
		token    string
		err      error

		remoteAddress = config.Context.RemoteAddress
	)

	fmt.Println(`Your email and password are your Basic Auth credentials.

Have you signed up with an authentication provider such as Google or GitHub?
Please, set up a WeDeploy password first at
` + color.Format(color.FgHiRed, "http://dashboard.%v/password/reset", remoteAddress) +
		"\nor you won't be able to continue.\n")
	if username, err = prompt.Prompt("Username"); err != nil {
		return err
	}

	if password, err = prompt.Hidden("Password"); err != nil {
		return err
	}

	token, err = loginserver.OAuthTokenFromBasicAuth(remoteAddress, username, password)

	if err != nil {
		return err
	}

	fmt.Println("")
	return saveUser(username, token)
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
Logout first with "we logout"`, config.Context.Username, config.Context.Remote, config.Context.RemoteAddress)
	}

	if noLaunchBrowser {
		println("OAuth implemented trough Basic Auth")
		return basicAuthLogin()
	}

	var service = &loginserver.Service{}
	var host, err = service.Listen(context.Background())

	if err != nil {
		return err
	}

	var loginURL = fmt.Sprintf("%s%s%s%s",
		defaults.DashboardURLPrefix,
		config.Context.RemoteAddress,
		"/login?redirect_uri=",
		url.QueryEscape(host))
	maybeOpenBrowser(loginURL)

	if err = service.Serve(); err != nil {
		return err
	}

	var username, token, tokenErr = service.Credentials()

	if tokenErr != nil {
		return tokenErr
	}

	return saveUser(username, token)
}

func saveUser(username, token string) (err error) {
	var g = config.Global
	var remote = g.Remotes[config.Context.Remote]
	remote.Username = username
	remote.Password = ""
	remote.Token = token

	g.Remotes[config.Context.Remote] = remote

	if err = g.Save(); err != nil {
		return err
	}

	fmt.Println(color.Format(color.FgHiCyan, `
You are now logged in as %s on %s
Go to http://wedeploy.com/docs/ to learn how to use WeDeploy`, username, remote.URL))
	return nil
}
