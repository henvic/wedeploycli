package cmdlogin

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/prompt"
)

// LoginCmd sets the user credential
var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Using Basic Authentication with your credentials",
	RunE:  loginRun,
}

func loginRun(cmd *cobra.Command, args []string) error {
	var (
		username string
		password string
		err      error
	)

	fmt.Println(`Your email and password are your credentials.

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

	var g = config.Global

	g.Username = username
	g.Password = password
	var err = g.Save()

	if err == nil {
		fmt.Println("Authentication information saved.")
	}

	return err
}
