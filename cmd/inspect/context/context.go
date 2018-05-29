package inspectcontext

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/inspector"
)

// ContextCmd for "we inspect context"
var ContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Get context info",
	Args:  cobra.NoArgs,
	RunE:  runE,
}

var (
	directory      string
	format         string
	showTypeFields bool
)

func init() {
	ContextCmd.Flags().StringVar(&directory, "directory", "", "Run the command on another directory")
	ContextCmd.Flags().StringVarP(&format, "format", "f", "", "Format the output using the given go template")
	ContextCmd.Flags().BoolVar(&showTypeFields, "fields", false, "Show type field names")
}

func runE(cmd *cobra.Command, args []string) (err error) {
	if directory == "" {
		directory = "."
	}

	if directory, err = filepath.Abs(directory); err != nil {
		return errwrap.Wrapf("Can not resolve directory: {{err}}", err)
	}

	if showTypeFields && format != "" {
		return errors.New("Incompatible use: --fields and --format can not be used together")
	}

	if showTypeFields {
		var i = inspector.ContextOverview{}
		fmt.Println(strings.Join(inspector.GetSpec(i), "\n"))
		return nil
	}

	var inspectMsg, inspectErr = inspector.InspectContext(format, directory)

	if inspectErr != nil {
		return inspectErr
	}

	fmt.Printf("%v\n", inspectMsg)
	return nil
}
