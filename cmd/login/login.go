package cmdlogin

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
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
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    loginRun,
}

func init() {
	LoginCmd.Flags().BoolVar(&noLaunchBrowser, "no-launch-browser", false, "Do not launch browser for authentication")
}

func maybeOpenBrowser(loginURL string) {
	switch errBrowser := browser.OpenURL(loginURL); {
	case errBrowser != nil:
		verbose.Debug("Can not open browser: " + errBrowser.Error())
		fmt.Println("Authenticate the CLI tool by visiting:")
	default:
		fmt.Println("Your browser has been opened to visit:")
	}

	fmt.Fprintf(os.Stdout, `%v

Waiting authentication via browser
`,
		loginURL)
}

func basicAuthLogin() error {
	var (
		username string
		password string
		token    string
		err      error
	)

	fmt.Println(`Your email and password are your Basic Auth credentials.

Have you signed up with an authentication provider such as Google or GitHub?
Please, set up a WeDeploy password first at
` + color.Format(color.FgHiRed, "http://dashboard.wedeploy.com/password/reset") +
		"\nor you won't be able to continue.\n")
	if username, err = prompt.Prompt("Username"); err != nil {
		return err
	}

	if password, err = prompt.Hidden("Password"); err != nil {
		return err
	}

	token, err = loginserver.OAuthTokenFromBasicAuth(username, password)

	if err != nil {
		return err
	}

	fmt.Println("")
	return saveUser(username, token)
}

func loginRun(cmd *cobra.Command, args []string) error {
	if noLaunchBrowser {
		println("OAuth implemented trough Basic Auth")
		return basicAuthLogin()
	}

	var service = &loginserver.Service{}
	var host, err = service.Listen(context.Background())

	if err != nil {
		return err
	}

	var loginURL = defaults.Dashboard + "/login?redirect_uri=" + url.QueryEscape(host)
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

	g.Username = username
	g.Password = ""
	g.Token = token

	if err = g.Save(); err != nil {
		return err
	}

	fmt.Println(color.Format(color.FgHiCyan, `
You are now logged in as %s
Go to http://wedeploy.com/docs/ to learn how to use WeDeploy`, g.Username))
	return nil
}
