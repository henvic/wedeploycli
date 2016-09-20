package cmdlogout

import (
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/config"
)

// LogoutCmd unsets the user credential
var LogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Revoke credentials",
	RunE:  logoutRun,
}

func logoutRun(cmd *cobra.Command, args []string) error {
	var g = config.Global

	g.Username = ""
	g.Password = ""
	g.Token = ""
	return g.Save()
}
