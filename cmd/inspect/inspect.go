package cmdinspect

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/inspector"
	"github.com/wedeploy/cli/projects"
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

	var inspectMsg, inspectErr = inspect(args[0])

	if inspectErr != nil {
		return inspectErr
	}

	fmt.Fprintf(os.Stdout, "%v\n", inspectMsg)
	return nil
}

func inspect(field string) (string, error) {
	switch field {
	case "context":
		return inspector.InspectContext(format, directory)
	case "project":
		return inspector.InspectProject(format, directory)
	case "container":
		return inspector.InspectContainer(format, directory)
	default:
		return "", fmt.Errorf(`Inspecting "%v" is not implemented.`, field)
	}
}

func printTypeFieldsSpec(field string) error {
	var i interface{}
	switch field {
	case "context":
		i = inspector.ContextOverview{}
	case "project":
		i = projects.Project{}
	case "container":
		i = containers.Container{}
	}

	if i == nil {
		return fmt.Errorf(`Spec for "%v" is not implemented.`, field)
	}

	fmt.Println(strings.Join(inspector.GetSpec(i), "\n"))
	return nil
}
