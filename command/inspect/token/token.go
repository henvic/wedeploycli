package token

import (
	"context"
	"errors"
	"fmt"

	"github.com/henvic/wedeploycli/cmdflagsfromhost"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/templates"
	"github.com/henvic/wedeploycli/usertoken"
	"github.com/spf13/cobra"
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
	var conf = wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes
	var remote = rl.Get(setupHost.Remote())

	if remote.Token == "" {
		return errors.New("user is not logged in")
	}

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
