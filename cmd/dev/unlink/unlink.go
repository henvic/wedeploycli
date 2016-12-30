package cmddevunlink

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
)

var (
	quiet        bool
	stopUnlinker = &unlinker{}
)

// StopCmd is the stop command to unlink a project or container
var StopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop a project or container",
	PreRunE: stopUnlinker.PreRun,
	RunE:    stopUnlinker.Run,
	Example: `  we dev stop
  we dev stop --project chat
  we dev stop --project chat --container data
  we dev stop --container data`,
}

type unlinker struct {
	project   string
	container string
	watcher   *list.Watcher
	list      *list.List
	end       bool
	err       error
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern:               cmdflagsfromhost.ProjectAndContainerPattern,
	UseProjectDirectory:   true,
	UseContainerDirectory: true,
	Requires: cmdflagsfromhost.Requires{
		Local: true,
	},
}

func init() {
	StopCmd.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"Unlink without watching status")
	setupHost.Init(StopCmd)
}

func (u *unlinker) PreRun(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("Invalid number of arguments.")
	}

	return setupHost.Process(args)
}

func (u *unlinker) Run(cmd *cobra.Command, args []string) error {
	var project = setupHost.Project()
	var container = setupHost.Container()

	u.project = project
	u.container = container

	if err := u.checkProjectOrContainerExists(); err != nil {
		return err
	}

	if quiet {
		u.do()
		return nil
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

func (u *unlinker) do() {
	switch u.container {
	case "":
		u.err = projects.Unlink(context.Background(), u.project)
	default:
		u.err = containers.Unlink(context.Background(), u.project, u.container)
	}

	u.end = true
}

func (u *unlinker) isDone() bool {
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

func (u *unlinker) handleWatchRequestError(err error) string {
	var ae, ok = err.(*apihelper.APIFault)

	if !ok || ae.Code != 404 {
		fmt.Fprintf(os.Stderr, "%v", errorhandling.Handle(err))
	}

	return "Unlinked successfully\n"
}

func (u *unlinker) watch() {
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

func (u *unlinker) checkProjectOrContainerExists() error {
	var err error
	if u.container == "" {
		_, err = projects.Get(context.Background(), u.project)
	} else {
		_, err = containers.Get(context.Background(), u.project, u.container)
	}

	return err
}
