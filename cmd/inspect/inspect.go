package cmdinspect

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/inspector"
	"github.com/wedeploy/cli/projects"
)

// InspectCmd returns information about current environment
var InspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect environment info",
	Long: `Use "we inspect" to peek inside a project or a service on your file system.
<type> = context | project | service`,
	Hidden:  true,
	PreRunE: cmdargslen.ValidateCmd(0, 1),
	RunE:    inspectRun,
	Example: `  we inspect context
  we inspect project
  we inspect service --format "{{.ID}}"`,
}

var (
	directory      string
	format         string
	showTypeFields bool
)

func init() {
	InspectCmd.Flags().StringVar(&directory, "directory", "", "Run the command on another directory")
	InspectCmd.Flags().StringVarP(&format, "format", "f", "", "Format the output using the given go template")
	InspectCmd.Flags().BoolVar(&showTypeFields, "fields", false, "Show type field names")
}

func inspectRun(cmd *cobra.Command, args []string) error {
	if directory == "" {
		directory = "."
	}

	var err error
	if directory, err = filepath.Abs(directory); err != nil {
		return errwrap.Wrapf("Can not resolve directory: {{err}}", err)
	}

	if len(args) != 1 {
		return errors.New("Expected: we inspect [context|project|service]")
	}

	if showTypeFields && format != "" {
		return errors.New("Incompatible use: --fields and --format can not be used together")
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
	case "service":
		return inspector.InspectService(format, directory)
	default:
		return "", fmt.Errorf(`inspecting "%v" is not implemented`, field)
	}
}

func printTypeFieldsSpec(field string) error {
	var i interface{}
	switch field {
	case "context":
		i = inspector.ContextOverview{}
	case "project":
		i = projects.Project{}
	case "service":
		i = services.Service{}
	}

	if i == nil {
		return fmt.Errorf(`spec for "%v" is not implemented`, field)
	}

	fmt.Println(strings.Join(inspector.GetSpec(i), "\n"))
	return nil
}
