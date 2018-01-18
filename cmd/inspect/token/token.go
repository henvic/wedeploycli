package token

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
)

// TokenCmd gets the user credential
var TokenCmd = &cobra.Command{
	Use:     "token",
	Short:   "Get current logged-in user token",
	PreRunE: preRun,
	RunE:    tokenRun,
	Args:    cobra.NoArgs,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.RemotePattern,
}

func init() {
	setupHost.Init(TokenCmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func tokenRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var config = wectx.Config()
	var remote = config.Remotes[setupHost.Remote()]
	fmt.Println(remote.Token)
	return nil
}
