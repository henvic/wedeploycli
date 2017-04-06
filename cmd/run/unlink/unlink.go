package cmdunlink

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmdargslen"
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
	Example: `  we run stop
  we run stop --project chat
  we run stop --project chat --container data
  we run stop --container data`,
}

type unlinker struct {
	project   string
	container string
	list      *list.List
	end       bool
	endMutex  sync.Mutex
	err       chan error
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern:               cmdflagsfromhost.ProjectAndContainerPattern,
	UseProjectDirectory:   true,
	UseContainerDirectory: true,
	Requires: cmdflagsfromhost.Requires{
		Project: true,
		Local:   true,
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

	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process()
}

func (u *unlinker) Run(cmd *cobra.Command, args []string) error {
	var project = setupHost.Project()
	var container = setupHost.Container()

	u.project = project
	u.container = container
	u.err = make(chan error, 1)

	if err := u.checkProjectOrContainerExists(); err != nil {
		return err
	}

	if quiet {
		u.do()
		return <-u.err
	}

	var queue sync.WaitGroup
	queue.Add(1)
	go u.do()
	go u.watch(queue.Done)
	queue.Wait()

	return <-u.err
}

func (u *unlinker) do() {
	switch u.container {
	case "":
		u.err <- projects.Unlink(context.Background(), u.project)
	default:
		u.err <- containers.Unlink(context.Background(), u.project, u.container)
	}

	u.endMutex.Lock()
	u.end = true
	u.endMutex.Unlock()
}

func (u *unlinker) getAddress() string {
	var address = fmt.Sprintf("%v.wedeploy.me", u.project)

	if u.container != "" {
		address = u.container + "." + address
	}

	return address
}

func (u *unlinker) isDone() bool {
	u.endMutex.Lock()
	var end = u.end
	u.endMutex.Unlock()

	if !end {
		return false
	}

	if len(u.list.Projects) == 0 {
		return true
	}

	var p = u.list.Projects[0]
	var c, e = p.Services(context.Background())

	if e != nil {
		var eaf, ok = e.(*apihelper.APIFault)
		return ok && eaf.Code == http.StatusNotFound
	}

	var _, ec = c.Get(u.container)
	return u.container != "" && ec != nil
}

func (u *unlinker) handleWatchRequestError(err error) string {
	var ae, ok = err.(*apihelper.APIFault)

	if !ok || !ae.Has("projectNotFound") {
		fmt.Fprintf(os.Stderr, "%v", errorhandling.Handle(err))
	}

	return u.getAddress() + " is shutdown\n"
}

func (u *unlinker) watch(done func()) {
	var filter = list.Filter{}

	filter.Project = u.project

	if u.container != "" {
		filter.Containers = []string{u.container}
	}
	u.list = list.New(filter)
	u.list.HandleRequestError = u.handleWatchRequestError
	u.list.StopCondition = u.isDone
	u.list.Start()
	done()
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
