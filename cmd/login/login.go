package cmdlogin

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
	RunE:  loginRun,
}

func loginRun(cmd *cobra.Command, args []string) error {
	var username = prompt.Prompt("Username")
	var password = prompt.Prompt("Password")
	var g = config.Global

	g.Username = username
	g.Password = password
	var err = g.Save()

	if err == nil {
		fmt.Println("Authentication information saved.")
	}

	return err
}
