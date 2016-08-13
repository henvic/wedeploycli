package cmdrestart

import (
	"sync"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/verbose"
)

// RestartCmd is used for getting restart
var RestartCmd = &cobra.Command{
	Use:   "restart [project] [container]",
	Short: "Restart project or container running on WeDeploy",
	RunE:  restartRun,
	Example: `we restart portal
we restart portal email`,
}

var quiet bool

func init() {
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
	err       error
	end       bool
}

func (r *restart) do() {
	switch r.container {
	case "":
		r.err = projects.Restart(r.project)
	default:
		r.err = containers.Restart(r.project, r.container)
	}

	r.end = true
}

func (r *restart) isDone() bool {
	if !r.end {
		return false
	}

	if len(r.list.Projects) == 0 {
		verbose.Debug("Unexpected behavior: no projects found.")
		return false
	}

	if r.container == "" && (r.list.Projects[0]).Health == "up" {
		return true
	}

	c, ok := r.list.Projects[0].Containers[r.container]

	if !ok {
		verbose.Debug("Unexpected behavior: no container found.")
		return false
	}

	return c.Health == "up"
}

func (r *restart) checkProjectOrContainerExists() error {
	var err error
	if r.container == "" {
		_, err = projects.Get(r.project)
	} else {
		_, err = containers.Get(r.project, r.container)
	}

	return err
}

func restartRun(cmd *cobra.Command, args []string) error {
	project, container, err := cmdcontext.GetProjectOrContainerID(args)

	if err != nil {
		return err
	}

	var r = &restart{
		project:   project,
		container: container,
	}

	if err = r.checkProjectOrContainerExists(); err != nil {
		return err
	}

	if quiet {
		r.do()
		return r.err
	}

	var queue sync.WaitGroup

	queue.Add(1)

	go func() {
		r.do()
	}()

	go func() {
		r.watch()
		queue.Done()
	}()

	queue.Wait()
	return r.err
}

func (r *restart) watch() {
	var filter = list.Filter{}

	filter.Project = r.project

	if r.container != "" {
		filter.Containers = []string{r.container}
	}

	r.list = list.New(filter)

	var watcher = list.NewWatcher(r.list)
	watcher.StopCondition = r.isDone
	watcher.Start()
}
