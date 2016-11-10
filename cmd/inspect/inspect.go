package cmdinspect

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/templates"
)

// InspectCmd returns information about current environment
var InspectCmd = &cobra.Command{
	Use:   "inspect <type> --format <format>",
	Short: "Inspect environment info",
	Long: `Use "we inspect" to peek inside a project or a container on your file system.
<type> = project | container`,
	RunE: inspectRun,
	Example: `we inspect project
we inspect project --fields
we inspect project --format "{{.ID}}"
we inspect container
we inspect container --fields
we inspect container --format "{{.Type}}"`,
}

var (
	format         string
	showTypeFields bool
)

func init() {
	InspectCmd.Flags().StringVarP(&format, "format", "f", "", "Format the output using the given go template")
	InspectCmd.Flags().BoolVar(&showTypeFields, "fields", false, "Show type field names")
}

func inspectRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("Wrong number of arguments; expected: we inspect <type>")
	}

	if showTypeFields && format != "" {
		return errors.New("Incompatible use: --fields and --format can't be used together")
	}

	switch args[0] {
	case "project":
		return inspectProject()
	case "container":
		return inspectContainer()
	default:
		return fmt.Errorf(`Inspecting "%v" is not implemented.`, args[0])
	}
}

func printTypeFieldNames(t interface{}) {
	val := reflect.ValueOf(t)
	for i := 0; i < val.NumField(); i += 1 {
		field := val.Type().Field(i)
		fmt.Println(field.Name + " " + field.Type.String())
	}
}

func inspectProject() error {
	if showTypeFields {
		printTypeFieldNames(projects.Project{})
		return nil
	}

	if config.Context.ProjectRoot == "" {
		return errors.New("Inspection failure: not inside project context.")
	}

	var project, err = projects.Read(config.Context.ProjectRoot)

	if err != nil {
		return errwrap.Wrapf("Inspection failure: {{err}}", err)
	}

	var content, eerr = templates.ExecuteOrList(format, project)

	if eerr != nil {
		return eerr
	}

	fmt.Println(content)
	return nil
}

func inspectContainer() error {
	if showTypeFields {
		printTypeFieldNames(containers.Container{})
		return nil
	}

	if config.Context.ContainerRoot == "" {
		return errors.New("Inspection failure: not inside container context.")
	}

	var container, err = containers.Read(config.Context.ContainerRoot)

	if err != nil {
		return errwrap.Wrapf("Inspection failure: {{err}}", err)
	}

	var content, eerr = templates.ExecuteOrList(format, container)

	if eerr != nil {
		return eerr
	}

	fmt.Println(content)
	return nil
}
