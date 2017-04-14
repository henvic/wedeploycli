package cmdlogout

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/userhome"
)

// LogoutCmd unsets the user credential
var LogoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Revoke credentials\n",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    logoutRun,
}

var rmConfig bool

func init() {
	LogoutCmd.Flags().BoolVar(&rmConfig, "rm-config-file", false, "Remove configuration file")
}

func logoutRun(cmd *cobra.Command, args []string) error {
	if rmConfig {
		return os.Remove(filepath.Join(userhome.GetHomeDir(), ".we"))
	}

	var g = config.Global

	g.Username = ""
	g.Password = ""
	g.Token = ""
	return g.Save()
}
