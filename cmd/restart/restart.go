package cmdrestart

import (
	"context"
	"sync"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
)

// RestartCmd is used for getting restart
var RestartCmd = &cobra.Command{
	Use:     "restart",
	Short:   "Restart project or container",
	PreRunE: preRun,
	RunE:    restartRun,
	Example: `  we restart --project chat --container data
  we restart --container data
  we restart --project chat --container data --remote local
  we restart --url data-chat.wedeploy.me`,
}

var quiet bool

var setupHost = cmdflagsfromhost.SetupHost{
	Requires: cmdflagsfromhost.Requires{
		Auth: true,
	},
	Pattern:               cmdflagsfromhost.FullHostPattern,
	UseProjectDirectory:   true,
	UseContainerDirectory: true,
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
	container string
	list      *list.List
	rwl       list.RestartWatchList
	err       error
	end       bool
	endMutex  sync.RWMutex
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (r *restart) do() {
	var err error

	switch r.container {
	case "":
		err = projects.Restart(r.ctx, r.project)
	default:
		err = containers.Restart(r.ctx, r.project, r.container)
	}

	r.endMutex.Lock()
	r.err = err
	r.end = true
	r.endMutex.Unlock()
	r.ctxCancel()
}

func (r *restart) checkProjectOrContainerExists() error {
	var p, err = projects.Get(context.Background(), r.project)
	r.rwl.SetInitialProjectHealthUID(p.HealthUID)

	switch {
	case err != nil:
		return err
	case r.container == "":
		return r.getContainerListForProjectRestart(p)
	default:
		return r.checkContainerExists()
	}
}

func (r *restart) getContainerListForProjectRestart(p projects.Project) error {
	var services, err = p.Services(context.Background())

	if err != nil {
		return err
	}

	var m = map[string]string{}

	for _, s := range services {
		m[s.ServiceID] = s.HealthUID
	}

	r.rwl.SetInitialContainersHealthUID(m)

	return nil
}

func (r *restart) checkContainerExists() error {
	var c, err = containers.Get(context.Background(), r.project, r.container)
	r.rwl.SetInitialContainersHealthUID(map[string]string{
		r.container: c.HealthUID,
	})
	return err
}

func (r *restart) watch() {
	r.rwl.Project = r.project
	r.rwl.IsStillRunning = r.hasRestartFinished

	if r.container != "" {
		r.rwl.Containers = []string{r.container}
	}

	r.rwl.Watch()
}

func (r *restart) hasRestartFinished() bool {
	r.endMutex.Lock()
	defer r.endMutex.Unlock()
	return r.end
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process()
}

func restartRun(cmd *cobra.Command, args []string) error {
	var ctx, cancel = context.WithCancel(context.Background())

	var r = &restart{
		project:   setupHost.Project(),
		container: setupHost.Container(),
		ctx:       ctx,
		ctxCancel: cancel,
	}

	if err := r.checkProjectOrContainerExists(); err != nil {
		return err
	}

	go r.do()

	if !quiet {
		r.watch()
	}

	<-r.ctx.Done()
	return r.err
}
