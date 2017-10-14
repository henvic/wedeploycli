package who

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdargslen"
)

// WhoCmd get the current user
var WhoCmd = &cobra.Command{
	Use:     "who",
	Short:   "Get who is using WeDeploy locally",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    whoRun,
}

func whoRun(cmd *cobra.Command, args []string) error {
	var wectx = we.Context()
	if wectx.Username() != "" {
		fmt.Printf("%s in %s (%s)\n",
			wectx.Username(),
			wectx.Remote(),
			wectx.InfrastructureDomain())
		return nil
	}

	return errors.New("User is not available")
}

func init() {
	WhoCmd.Hidden = true
}
