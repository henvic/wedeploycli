package cmdlogout

import (

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/config"
)

// LogoutCmd unsets the user credential
var LogoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Logout from your account",
	PreRunE: preRun,
	RunE:    logoutRun,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.RemotePattern,
}

func init() {
	setupHost.Init(LogoutCmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process()
}

func logoutRun(cmd *cobra.Command, args []string) error {
	var g = config.Global
	var remote = g.Remotes[config.Context.Remote]

	remote.Username = ""
	remote.Password = ""
	remote.Token = ""
	g.Remotes.Set(config.Context.Remote, remote)
	return g.Save()
}
