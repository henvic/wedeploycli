package cmdauth

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/prompt"
)

// LoginCmd sets the user credential
var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Using Basic Authentication with your credentials",
	Run:   loginRun,
}

// LogoutCmd unsets the user credential
var LogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Revoke credentials",
	Run:   logoutRun,
}

func loginRun(cmd *cobra.Command, args []string) {
	var username = prompt.Prompt("Username")
	var password = prompt.Prompt("Password")
	var g = config.Global

	g.Username = username
	g.Password = password
	g.Save()

	fmt.Println("Authentication information saved.")
}

func logoutRun(cmd *cobra.Command, args []string) {
	var g = config.Global

	g.Username = ""
	g.Password = ""
	g.Token = ""
	g.Save()
}
