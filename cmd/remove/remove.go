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
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
)

var (
	quiet bool
)

// RemoveCmd is the remove command to undeploy a project or service
var RemoveCmd = &cobra.Command{
	Use:     "remove",
	Short:   "Delete project or services",
	PreRunE: preRun,
	RunE:    run,
	Example: `  we remove --url data.chat.wedeploy.me
  we remove --project chat
  we remove --project chat --service data`,
}

type undeployer struct {
	context              context.Context
	project              string
	service              string
	infrastructureDomain string
	serviceDomain        string
	list                 *list.List
	end                  bool
	endMutex             sync.Mutex
	err                  chan error
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern:             cmdflagsfromhost.FullHostPattern,
	UseServiceDirectory: true,
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
		return errors.New("invalid number of arguments")
	}

	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process()
}

func run(cmd *cobra.Command, args []string) error {
	var u = undeployer{
		context:              context.Background(),
		project:              setupHost.Project(),
		service:              setupHost.Service(),
		infrastructureDomain: setupHost.InfrastructureDomain(),
		serviceDomain:        setupHost.ServiceDomain(),
		err:                  make(chan error, 1),
	}

	if err := u.checkProjectOrServiceExists(); err != nil {
		return err
	}

	go u.do()

	if !quiet {
		u.watch()
	}

	return <-u.err
}

func (u *undeployer) do() {
	switch u.service {
	case "":
		u.err <- projects.Unlink(u.context, u.project)
	default:
		u.err <- services.Unlink(u.context, u.project, u.service)
	}

	u.endMutex.Lock()
	u.end = true
	u.endMutex.Unlock()
}

func (u *undeployer) getAddress() string {
	var address = fmt.Sprintf("%v.%v", u.project, u.serviceDomain)

	if u.service != "" {
		address = u.service + "." + address
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

	var _, ec = c.Get(u.service)
	return u.service != "" && ec != nil
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

	if u.service != "" {
		filter.Services = []string{u.service}
	}

	u.list = list.New(filter)
	u.list.HandleRequestError = u.handleWatchRequestError
	u.list.StopCondition = u.isDone
	u.list.Start()
}

func (u *undeployer) checkProjectOrServiceExists() (err error) {
	if u.service == "" {
		_, err = projects.Get(u.context, u.project)
	} else {
		_, err = services.Get(u.context, u.project, u.service)
	}

	return err
}
