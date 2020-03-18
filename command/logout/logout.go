package logout

import (
	"context"
	"fmt"

	"github.com/henvic/wedeploycli/cmdflagsfromhost"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/fancy"
	"github.com/spf13/cobra"
)

// LogoutCmd unsets the user credential
var LogoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Logout from your account\n\t\t",
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
	var params = conf.GetParams()
	var rl = params.Remotes
	var remote = rl.Get(wectx.Remote())

	switch remote.Username {
	case "":
		fmt.Println(fancy.Info(fmt.Sprintf(`You are not logged in on %s.`,
			remote.Infrastructure)))
	default:
		fmt.Printf("You (%s) have been logged out of %s.\n",
			remote.Username,
			remote.Infrastructure)
	}

	remote.Username = ""
	remote.Token = ""
	rl.Set(wectx.Remote(), remote)
	return conf.Save()
}
