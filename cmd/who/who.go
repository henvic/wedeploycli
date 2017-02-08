package cmdwho

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/config"
)

// WhoCmd get the current user
var WhoCmd = &cobra.Command{
	Use:     "who",
	Short:   "Get who is using WeDeploy locally",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    whoRun,
}

func whoRun(cmd *cobra.Command, args []string) error {
	var g = config.Global

	if g.Username != "" {
		fmt.Println(g.Username)
		return nil
	}

	return errors.New("User is not available")
}

func init() {
	WhoCmd.Hidden = true
}
