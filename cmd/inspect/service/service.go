package inspectservice

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/inspector"
	"github.com/wedeploy/cli/services"
)

// ServiceCmd for "we inspect service"
var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Get service info",
	Args:  cobra.NoArgs,
	RunE:  runE,
}

var (
	directory      string
	format         string
	showTypeFields bool
)

func init() {
	ServiceCmd.Flags().StringVar(&directory, "directory", "", "Run the command on another directory")
	ServiceCmd.Flags().StringVarP(&format, "format", "f", "", "Format the output using the given go template")
	ServiceCmd.Flags().BoolVar(&showTypeFields, "fields", false, "Show type field names")
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
		var i = services.ServicePackage{}
		fmt.Println(strings.Join(inspector.GetSpec(i), "\n"))
		return nil
	}

	var inspectMsg, inspectErr = inspector.InspectService(format, directory)

	if inspectErr != nil {
		return inspectErr
	}

	fmt.Fprintf(os.Stdout, "%v\n", inspectMsg)
	return nil
}
