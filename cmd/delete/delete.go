package cmddelete

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/cmd/canceled"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
)

var (
	quiet bool
	force bool
)

// DeleteCmd is the delete command to undeploy a project or service
var DeleteCmd = &cobra.Command{
	Use:     "delete",
	Hidden:  true,
	Short:   "Delete project or services",
	PreRunE: preRun,
	RunE:    run,
	Example: `  we delete --url data.chat.wedeploy.io
  we delete --project chat
  we delete --project chat --service data`,
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
	Pattern: cmdflagsfromhost.FullHostPattern,
	Requires: cmdflagsfromhost.Requires{
		Project: true,
		Auth:    true,
	},
}

func init() {
	DeleteCmd.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"Undeploy services without watching status")
	DeleteCmd.Flags().BoolVar(&force, "force", false,
		"Force deleting services without confirmation")
	setupHost.Init(DeleteCmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		return errors.New("invalid number of arguments")
	}

	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process(we.Context())
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

	if err := confirmation(); err != nil {
		return err
	}

	go u.do()

	if !quiet {
		u.watch()
	}

	return <-u.err
}

func confirmation() error {
	if force {
		return nil
	}

	var question string

	if setupHost.Service() != "" {
		question = fmt.Sprintf(`Do you really want to delete the service "%v" on project "%v"?`,
			color.Format(color.Bold, setupHost.Service()),
			color.Format(color.Bold, setupHost.Project()))
	} else {
		question = fmt.Sprintf(`Do you really want to delete the project "%v"?`,
			color.Format(color.Bold, setupHost.Project()))
	}

	var confirm, askErr = fancy.Boolean(question)

	if askErr != nil {
		return askErr
	}

	if confirm {
		return nil
	}

	return canceled.CancelCommand("delete canceled")
}

func (u *undeployer) do() {
	switch u.service {
	case "":
		projectsClient := projects.New(we.Context())
		u.err <- projectsClient.Unlink(u.context, u.project)
	default:
		servicesClient := services.New(we.Context())
		u.err <- servicesClient.Unlink(u.context, u.project, u.service)
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
	var c, e = p.Services(u.context, services.New(we.Context()))

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
	u.list.Start(we.Context())
}

func (u *undeployer) checkProjectOrServiceExists() (err error) {
	if u.service == "" {
		projectsClient := projects.New(we.Context())
		_, err = projectsClient.Get(u.context, u.project)
	} else {
		servicesClient := services.New(we.Context())
		_, err = servicesClient.Get(u.context, u.project, u.service)
	}

	return err
}
