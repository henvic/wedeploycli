package inspectconfig

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/inspector"
)

// ConfigCmd for "lcp inspect config"
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Get config info",
	Args:  cobra.NoArgs,
	RunE:  runE,
}

var (
	format         string
	showTypeFields bool
)

func init() {
	ConfigCmd.Flags().StringVarP(&format, "format", "f", "", "Format the output using the given go template")
	ConfigCmd.Flags().BoolVar(&showTypeFields, "fields", false, "Show type field names")
}

func runE(cmd *cobra.Command, args []string) (err error) {
	if showTypeFields && format != "" {
		return errors.New("incompatible use: --fields and --format cannot be used together")
	}

	if showTypeFields {
		// TODO(henvic): consider to expose *remotes.List primitives
		var i = config.Params{}
		fmt.Println(strings.Join(inspector.GetSpec(i), "\n"))
		return nil
	}

	wectx := we.Context()

	var inspectMsg, inspectErr = inspector.InspectConfig(format, wectx)

	if inspectErr != nil {
		return inspectErr
	}

	fmt.Printf("%v\n", inspectMsg)
	return nil
}
