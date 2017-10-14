package restart

import (
	"context"
	"sync"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
)

// RestartCmd is used for getting restart
var RestartCmd = &cobra.Command{
	Use:     "restart",
	Short:   "Restart project or services",
	PreRunE: preRun,
	RunE:    restartRun,
	Example: `  we restart --project chat --service data
  we restart --service data
  we restart --project chat --service data --remote wedeploy
  we restart --url data-chat.wedeploy.io`,
}

var quiet bool

var setupHost = cmdflagsfromhost.SetupHost{
	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
	},
	Pattern: cmdflagsfromhost.FullHostPattern,
}

func init() {
	setupHost.Init(RestartCmd)
	RestartCmd.Flags().BoolVarP(
		&quiet,
		"quiet",
		"q",
		false,
		"Reset without watching status.")
}

type restart struct {
	project   string
	service   string
	rwl       list.RestartWatchList
	err       error
	end       bool
	endMutex  sync.RWMutex
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (r *restart) do() {
	var err error

	switch r.service {
	case "":
		projectsClient := projects.New(we.Context())
		err = projectsClient.Restart(r.ctx, r.project)
	default:
		servicesClient := services.New(we.Context())
		err = servicesClient.Restart(r.ctx, r.project, r.service)
	}

	r.endMutex.Lock()
	r.err = err
	r.end = true
	r.endMutex.Unlock()
	r.ctxCancel()
}

func (r *restart) checkProjectOrServiceExists() error {
	projectsClient := projects.New(we.Context())
	var p, err = projectsClient.Get(context.Background(), r.project)
	r.rwl.SetInitialProjectHealthUID(p.HealthUID)

	switch {
	case err != nil:
		return err
	case r.service == "":
		return r.getServiceListForProjectRestart(p)
	default:
		return r.checkServiceExists()
	}
}

func (r *restart) getServiceListForProjectRestart(p projects.Project) error {
	servicesClient := services.New(we.Context())
	var services, err = p.Services(context.Background(), servicesClient)

	if err != nil {
		return err
	}

	var m = map[string]string{}

	for _, s := range services {
		m[s.ServiceID] = s.HealthUID
	}

	r.rwl.SetInitialServicesHealthUID(m)

	return nil
}

func (r *restart) checkServiceExists() error {
	servicesClient := services.New(we.Context())
	var c, err = servicesClient.Get(context.Background(), r.project, r.service)
	r.rwl.SetInitialServicesHealthUID(map[string]string{
		r.service: c.HealthUID,
	})
	return err
}

func (r *restart) watch() {
	r.rwl.Project = r.project
	r.rwl.IsStillRunning = r.hasRestartFinished

	if r.service != "" {
		r.rwl.Services = []string{r.service}
	}

	r.rwl.Watch()
}

func (r *restart) hasRestartFinished() bool {
	r.endMutex.Lock()
	defer r.endMutex.Unlock()
	return r.end
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(we.Context())
}

func restartRun(cmd *cobra.Command, args []string) error {
	var r = &restart{
		project: setupHost.Project(),
		service: setupHost.Service(),
	}

	r.ctx, r.ctxCancel = context.WithCancel(
		context.Background())

	if err := r.checkProjectOrServiceExists(); err != nil {
		return err
	}

	go r.do()

	if !quiet {
		r.watch()
	}

	<-r.ctx.Done()
	return r.err
}
