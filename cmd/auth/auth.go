package auth

import (
	"fmt"

	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/defaults"
	"github.com/launchpad-project/cli/prompt"
	"github.com/spf13/cobra"
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
	var csg = config.Stores["global"]
	var username = prompt.Prompt("Username")
	var password = prompt.Prompt("Password")

	csg.Set("endpoint", defaults.Endpoint)
	csg.Set("username", username)
	csg.Set("password", password)
	csg.Save()

	fmt.Println("Authentication information saved.")
}

func logoutRun(cmd *cobra.Command, args []string) {
	var csg = config.Stores["global"]

	csg.Set("username", "")
	csg.Set("password", "")
	csg.Save()
}
