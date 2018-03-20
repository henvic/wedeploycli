package token

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/templates"
	"github.com/wedeploy/cli/usertoken"
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

var format string

func init() {
	TokenCmd.Flags().StringVarP(&format, "format", "f", "", "Format the output using the given go template")
	TokenCmd.Flag("format").Hidden = true
	setupHost.Init(TokenCmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func tokenRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	var config = wectx.Config()
	var remote = config.Remotes[setupHost.Remote()]

	if format == "" {
		fmt.Println(remote.Token)
		return nil
	}

	t, err := usertoken.ParseUnsignedJSONWebToken(remote.Token)

	if err != nil {
		return err
	}

	print, err := templates.ExecuteOrList(format, t)

	if err != nil {
		return err
	}

	fmt.Println(print)
	return nil
}
