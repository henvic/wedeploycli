package cmdremove

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
	quiet bool
)

// RemoveCmd is the remove command to undeploy a project or container
var RemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove a project or container",
	PreRunE: preRun,
	RunE:    run,
	Example: `  we remove
  we remove --project chat
  we remove --project chat --container data
  we remove --container data`,
}

type undeployer struct {
	context       context.Context
	project       string
	container     string
	remoteAddress string
	list          *list.List
	end           bool
	endMutex      sync.Mutex
	err           chan error
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern:               cmdflagsfromhost.FullHostPattern,
	UseProjectDirectory:   true,
	UseContainerDirectory: true,
	Requires: cmdflagsfromhost.Requires{
		Project: true,
		Auth:    true,
	},
}

func init() {
	RemoveCmd.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"undeploy without watching status")
	setupHost.Init(RemoveCmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("Invalid number of arguments.")
	}

	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process()
}

func run(cmd *cobra.Command, args []string) error {
	var u = undeployer{
		context:       context.Background(),
		project:       setupHost.Project(),
		container:     setupHost.Container(),
		remoteAddress: setupHost.RemoteAddress(),
		err:           make(chan error, 1),
	}

	if err := u.checkProjectOrContainerExists(); err != nil {
		return err
	}

	go u.do()

	if !quiet {
		u.watch()
	}

	return <-u.err
}

func (u *undeployer) do() {
	switch u.container {
	case "":
		u.err <- projects.Unlink(u.context, u.project)
	default:
		u.err <- containers.Unlink(u.context, u.project, u.container)
	}

	u.endMutex.Lock()
	u.end = true
	u.endMutex.Unlock()
}

func (u *undeployer) getAddress() string {
	var address = fmt.Sprintf("%v.%v", u.project, u.remoteAddress)

	if u.container != "" {
		address = u.container + "." + address
	}

	return address
}

func (u *undeployer) isDone() bool {
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
	var c, e = p.Services(u.context)

	if e != nil {
		var eaf, ok = e.(*apihelper.APIFault)
		return ok && eaf.Status == http.StatusNotFound
	}

	var _, ec = c.Get(u.container)
	return u.container != "" && ec != nil
}

func (u *undeployer) handleWatchRequestError(err error) string {
	var ae, ok = err.(*apihelper.APIFault)

	if !ok || !ae.Has("projectNotFound") {
		fmt.Fprintf(os.Stderr, "%v", errorhandling.Handle(err))
	}

	return u.getAddress() + " is shutdown\n"
}

func (u *undeployer) watch() {
	var queue sync.WaitGroup
	queue.Add(1)
	go func() {
		u.watchRoutine()
		queue.Done()
	}()
	queue.Wait()
}

func (u *undeployer) watchRoutine() {
	var filter = list.Filter{}

	filter.Project = u.project

	if u.container != "" {
		filter.Containers = []string{u.container}
	}

	u.list = list.New(filter)
	u.list.HandleRequestError = u.handleWatchRequestError
	u.list.StopCondition = u.isDone
	u.list.Start()
}

func (u *undeployer) checkProjectOrContainerExists() (err error) {
	if u.container == "" {
		_, err = projects.Get(u.context, u.project)
	} else {
		_, err = containers.Get(u.context, u.project, u.container)
	}

	return err
}
