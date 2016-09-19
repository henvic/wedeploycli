package cmdcreate

import (
	"errors"
	"path/filepath"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/createctx"
)

var (
	project   string
	container string
)

// CreateCmd creates a project or container
var CreateCmd = &cobra.Command{
	Use:   "create --project <project> --container <container> [directory]",
	Short: "Creates a project or container",
	Long: `Use "we create" to create projects and containers.
You can create a project anywhere on your machine.
Containers can only be created from inside projects and
are stored on the first subdirectory of its project.`,
	RunE:    createRun,
	Example: `we create --project cinema --container projector room`,
}

func init() {
	CreateCmd.Flags().StringVar(&project, "project", "", "Project ID")
	CreateCmd.Flags().StringVar(&container, "container", "", "Container ID")
}

func createRun(cmd *cobra.Command, args []string) error {
	var directory = ""

	switch len(args) {
	case 0:
	case 1:
		directory = args[0]
	default:
		return errors.New("Invalid number of arguments")
	}

	if directory == "" {
		directory = "."
	}

	directory, err := filepath.Abs(directory)

	if err != nil {
		return errwrap.Wrapf("Can't get absolute path: {{err}}", err)
	}

	if project == "" {
		return errors.New("Missing --project <id>")
	}

	if container == "" {
		return createctx.NewProject(project, directory)
	}

	return createctx.NewContainer(project, container, directory)
}
