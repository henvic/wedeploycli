package inspectservice

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/wedeploycli/inspector"
	"github.com/henvic/wedeploycli/services"
	"github.com/spf13/cobra"
)

// ServiceCmd for "lcp inspect service"
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
		return errwrap.Wrapf("can't resolve directory: {{err}}", err)
	}

	if showTypeFields && format != "" {
		return errors.New("incompatible use: --fields and --format cannot be used together")
	}

	if showTypeFields {
		var p = services.Package{}
		fmt.Println(strings.Join(inspector.GetSpec(p), "\n"))
		return nil
	}

	var inspectMsg, inspectErr = inspector.InspectService(format, directory)

	if inspectErr != nil {
		return inspectErr
	}

	fmt.Printf("%v\n", inspectMsg)
	return nil
}
