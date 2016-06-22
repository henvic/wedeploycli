package cmdlink

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/link"
)

// LinkCmd links the given project or container locally
var LinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Links the given project or container locally",
	Run:   linkRun,
	Example: `we link
we link <project>
we link <container>`,
}

func getContainersFromScope() []string {
	if config.Context.ContainerRoot != "" {
		_, container := filepath.Split(config.Context.ContainerRoot)
		return []string{container}
	}

	var list, err = containers.GetListFromDirectory(config.Context.ProjectRoot)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return list
}

func linkContainersFeedback(success []string, err error) {
	for _, s := range success {
		fmt.Println(s)
	}

	if len(success) != 0 && err != nil {
		fmt.Println("")
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func linkRun(cmd *cobra.Command, args []string) {
	// calling it for side-effects
	getProjectOrContainerID(args)

	var success, err = link.All(config.Context.ProjectRoot,
		getContainersFromScope())

	linkContainersFeedback(success, err)
}

func getProjectOrContainerID(args []string) (string, string) {
	var project, container, err = cmdcontext.GetProjectOrContainerID(args)

	if err != nil {
		println("fatal: not a project")
		os.Exit(1)
	}

	return project, container
}
