package inspect

import (
	inspectconfig "github.com/henvic/wedeploycli/command/inspect/config"
	inspectcontext "github.com/henvic/wedeploycli/command/inspect/context"
	inspectservice "github.com/henvic/wedeploycli/command/inspect/service"
	"github.com/henvic/wedeploycli/command/inspect/token"
	"github.com/spf13/cobra"
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
