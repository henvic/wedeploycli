package inspect

import (
	"github.com/spf13/cobra"
	inspectconfig "github.com/wedeploy/cli/command/inspect/config"
	inspectcontext "github.com/wedeploy/cli/command/inspect/context"
	inspectservice "github.com/wedeploy/cli/command/inspect/service"
	"github.com/wedeploy/cli/command/inspect/token"
)

// InspectCmd returns information about current environment
var InspectCmd = &cobra.Command{
	Use:    "inspect",
	Short:  "Inspect environment info",
	Hidden: true,
	Args:   cobra.MaximumNArgs(1),
	Example: `  lcp inspect context
  lcp inspect service --format "{{.ID}}"`,
}

func init() {
	InspectCmd.AddCommand(inspectservice.ServiceCmd)
	InspectCmd.AddCommand(inspectcontext.ContextCmd)
	InspectCmd.AddCommand(inspectconfig.ConfigCmd)
	InspectCmd.AddCommand(token.TokenCmd)
}
