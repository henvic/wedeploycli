package inspect

import (
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/inspect/config"
	"github.com/wedeploy/cli/cmd/inspect/context"
	"github.com/wedeploy/cli/cmd/inspect/service"
	"github.com/wedeploy/cli/cmd/inspect/token"
)

// InspectCmd returns information about current environment
var InspectCmd = &cobra.Command{
	Use:    "inspect",
	Short:  "Inspect environment info",
	Hidden: true,
	Args:   cobra.MaximumNArgs(1),
	Example: `  we inspect context
  we inspect service --format "{{.ID}}"`,
}

func init() {
	InspectCmd.AddCommand(inspectservice.ServiceCmd)
	InspectCmd.AddCommand(inspectcontext.ContextCmd)
	InspectCmd.AddCommand(inspectconfig.ConfigCmd)
	InspectCmd.AddCommand(token.TokenCmd)
}
