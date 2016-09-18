package cmdunlink

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/errorhandling"
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
	watcher   *list.Watcher
	list      *list.List
	end       bool
	err       error
}

func (u *unlink) do() {
	switch u.container {
	case "":
		u.err = projects.Unlink(context.Background(), u.project)
	default:
		u.err = containers.Unlink(context.Background(), u.project, u.container)
	}

	u.end = true
}

func (u *unlink) isDone() bool {
	if !u.end {
		return false
	}

	if len(u.watcher.List.Projects) == 0 {
		return true
	}

	if u.container != "" && u.watcher.List.Projects[0].Containers[u.container] == nil {
		u.watcher.Teardown = func() string {
			return "Container unlinked successfully!\n"
		}

		return true
	}

	return false
}

func (u *unlink) handleWatchRequestError(err error) string {
	var ae, ok = err.(*apihelper.APIFault)

	if !ok || ae.Code != 404 {
		fmt.Fprintf(os.Stderr, "%v", errorhandling.Handle(err))
	}

	return "Unlinked successfully\n"
}

func (u *unlink) watch() {
	var filter = list.Filter{}

	filter.Project = u.project

	if u.container != "" {
		filter.Containers = []string{u.container}
	}
	u.watcher = list.NewWatcher(list.New(filter))
	u.watcher.List.HandleRequestError = u.handleWatchRequestError
	u.watcher.StopCondition = u.isDone
	u.watcher.Start()
}

func (u *unlink) checkProjectOrContainerExists() error {
	var err error
	if u.container == "" {
		_, err = projects.Get(context.Background(), u.project)
	} else {
		_, err = containers.Get(context.Background(), u.project, u.container)
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
		return err
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
