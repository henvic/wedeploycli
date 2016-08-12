package cmdlink

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/hashicorp/errwrap"
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
	RunE:  linkRun,
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

func getContainersDirectoriesFromScope() ([]string, error) {
	if config.Context.ContainerRoot != "" {
		_, container := filepath.Split(config.Context.ContainerRoot)
		return []string{container}, nil
	}

	var list, err = containers.GetListFromDirectory(config.Context.ProjectRoot)

	if err != nil {
		err = errwrap.Wrapf("Error retrieving containers list from directory: {{err}}", err)
	}

	return list, err
}

func linkRun(cmd *cobra.Command, args []string) error {
	if _, _, err := cmdcontext.GetProjectOrContainerID(args); err != nil {
		return nil
	}

	var csDirs, err = getContainersDirectoriesFromScope()

	if err != nil {
		return err
	}

	var m = &link.Machine{
		FErrStream: os.Stderr,
	}

	if err = m.Setup(config.Context.ProjectRoot, csDirs); err != nil {
		return err
	}

	if quiet {
		m.Run()
		return nil
	}

	var queue sync.WaitGroup

	queue.Add(1)

	go func() {
		m.Run()
	}()

	go func() {
		m.Watch()
		queue.Done()
	}()

	queue.Wait()

	if len(m.Errors.List) != 0 {
		return m.Errors
	} else {
		return nil
	}
}
