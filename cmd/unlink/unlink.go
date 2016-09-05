package cmdunlink

import (
	"sync"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
)

// UnlinkCmd unlinks the given project or container locally
var UnlinkCmd = &cobra.Command{
	Use:   "unlink",
	Short: "Unlinks the given project or container locally",
	RunE:  unlinkRun,
	Example: `we unlink
we unlink <project>
we unlink <project> <container>
we unlink <container>`,
}

var quiet bool

func init() {
	UnlinkCmd.Flags().BoolVarP(
		&quiet,
		"quiet",
		"q",
		false,
		"Unlink without watching status.")
}

type unlink struct {
	project   string
	container string
	list      *list.List
	end       bool
	err       error
}

func (u *unlink) do() {
	switch u.container {
	case "":
		u.err = projects.Unlink(u.project)
	default:
		u.err = containers.Unlink(u.project, u.container)
	}

	u.end = true
}

func (u *unlink) isDone() bool {
	if !u.end {
		return false
	}

	if len(u.list.Projects) == 0 {
		return true
	}

	if u.container != "" && u.list.Projects[0].Containers[u.container] == nil {
		return true
	}

	return false
}

func (u *unlink) watch() {
	var filter = list.Filter{}

	filter.Project = u.project

	if u.container != "" {
		filter.Containers = []string{u.container}
	}

	u.list = list.New(filter)

	u.list.StyledNotFound = true

	var watcher = list.NewWatcher(u.list)
	watcher.StopCondition = u.isDone
	watcher.Start()
}

func (u *unlink) checkProjectOrContainerExists() error {
	var err error
	if u.container == "" {
		_, err = projects.Get(u.project)
	} else {
		_, err = containers.Get(u.project, u.container)
	}

	return err
}

func handleCheckProjectOrContainerError(err error) error {
	switch err.(type) {
	case *apihelper.APIFault:
		var ae = err.(*apihelper.APIFault)

		if ae.Has("documentNotFound") {
			println("Successfully unlinked")
			return nil
		}
	}

	return err
}

func unlinkRun(cmd *cobra.Command, args []string) error {
	var project, container, err = cmdcontext.GetProjectOrContainerID(args)

	if err != nil {
		return errwrap.Wrapf("fatal: {{err}}", err)
	}

	var u = &unlink{
		project:   project,
		container: container,
	}

	if err = u.checkProjectOrContainerExists(); err != nil {
		return handleCheckProjectOrContainerError(err)
	}

	if quiet {
		u.do()
		return err
	}

	var queue sync.WaitGroup

	queue.Add(1)

	go func() {
		u.do()
	}()

	go func() {
		u.watch()
		queue.Done()
	}()

	queue.Wait()

	if u.err != nil {
		return u.err
	}

	return nil
}
