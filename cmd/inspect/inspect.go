package cmdinspect

import (
	"errors"
	"fmt"

	"path/filepath"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/inspector"
)

// InspectCmd returns information about current environment
var InspectCmd = &cobra.Command{
	Use:   "inspect <type> --format <format>",
	Short: "Inspect environment info",
	Long: `Use "we inspect" to peek inside a project or a container on your file system.
<type> = context | project | container`,
	RunE: inspectRun,
	Example: `we inspect context
we inspect context --fields
we inspect context --format "{{.Scope}}"
we inspect project
we inspect project --fields
we inspect project --format "{{.ID}}"
we inspect container
we inspect container --fields
we inspect container --format "{{.Type}}"`,
}

var (
	directory      string
	format         string
	showTypeFields bool
)

func init() {
	InspectCmd.Flags().StringVar(&directory, "directory", "", "Overrides current working directory")
	InspectCmd.Flags().StringVarP(&format, "format", "f", "", "Format the output using the given go template")
	InspectCmd.Flags().BoolVar(&showTypeFields, "fields", false, "Show type field names")
}

func inspectRun(cmd *cobra.Command, args []string) error {
	if directory == "" {
		directory = "."
	}

	var err error
	if directory, err = filepath.Abs(directory); err != nil {
		return errwrap.Wrapf("Can't resolve directory: {{err}}", err)
	}

	if len(args) != 1 {
		return errors.New("Wrong number of arguments; expected: we inspect <type>")
	}

	if showTypeFields && format != "" {
		return errors.New("Incompatible use: --fields and --format can't be used together")
	}

	if showTypeFields {
		return printTypeFieldsSpec(args[0])
	}

	return inspect(args[0])
}

func inspect(field string) error {
	switch field {
	case "context":
		return inspector.InspectContext(format, directory)
	case "project":
		return inspector.InspectProject(format, directory)
	case "container":
		return inspector.InspectContainer(format, directory)
	default:
		return fmt.Errorf(`Inspecting "%v" is not implemented.`, field)
	}
}

func printTypeFieldsSpec(field string) error {
	switch field {
	case "context":
		inspector.PrintContextSpec()
		return nil
	case "project":
		inspector.PrintProjectSpec()
		return nil
	case "container":
		inspector.PrintContainerSpec()
		return nil
	default:
		return fmt.Errorf(`Spec for "%v" is not implemented.`, field)
	}
}
