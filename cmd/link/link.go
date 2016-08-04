package cmdlink

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

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

var quiet bool

func init() {
	LinkCmd.Flags().BoolVarP(
		&quiet,
		"quiet",
		"q",
		false,
		"Link without watching status.")
}

func getContainersDirectoriesFromScope() []string {
	if config.Context.ContainerRoot != "" {
		_, container := filepath.Split(config.Context.ContainerRoot)
		return []string{container}
	}

	var list, err = containers.GetListFromDirectory(config.Context.ProjectRoot)

	if err != nil {
		fmt.Fprintln(os.Stderr, "Error retrieving containers list from directory.")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return list
}

func linkErrorsFeedback(err *link.Errors) {
	if len(err.List) != 0 {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func linkRun(cmd *cobra.Command, args []string) {
	// calling it for side-effects
	getProjectOrContainerID(args)
	var csDirs = getContainersDirectoriesFromScope()

	var m = &link.Machine{
		FErrStream: os.Stderr,
	}

	var err = m.Setup(config.Context.ProjectRoot, csDirs)

	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	if quiet {
		m.Run()
		return
	}

	var queue sync.WaitGroup

	queue.Add(2)

	go func() {
		m.Run()
		queue.Done()
	}()

	go func() {
		m.Watch()
		queue.Done()
	}()

	queue.Wait()
	linkErrorsFeedback(m.Errors)
}

func getProjectOrContainerID(args []string) (string, string) {
	var project, container, err = cmdcontext.GetProjectOrContainerID(args)

	if err != nil {
		println("fatal: not a project")
		os.Exit(1)
	}

	return project, container
}
