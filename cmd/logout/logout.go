package logout

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/fancy"
)

// LogoutCmd unsets the user credential
var LogoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Logout from your account",
	Args:    cobra.NoArgs,
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
	return setupHost.Process(context.Background(), we.Context())
}

func logoutRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var conf = wectx.Config()
	var remote = conf.Remotes[wectx.Remote()]

	switch remote.Username {
	case "":
		fmt.Println(fancy.Info(fmt.Sprintf(`You are not logged in on %s.`,
			remote.Infrastructure)))
	default:
		fmt.Println(fancy.Success(fmt.Sprintf("You (%s) have been logged out of %s.",
			remote.Username,
			remote.Infrastructure)))
	}

	remote.Username = ""
	remote.Token = ""
	conf.Remotes.Set(wectx.Remote(), remote)
	return conf.Save()
}
