package cmdrestart

import (
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdcontext"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
)

// RestartCmd is used for getting restart
var RestartCmd = &cobra.Command{
	Use:   "restart [project] [container]",
	Short: "Restart project or container running on WeDeploy",
	Run:   restartRun,
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
	end       bool
}

func (r *restart) do() {
	switch r.container {
	case "":
		projects.Restart(r.project)
	default:
		containers.Restart(r.project, r.container)
	}

	r.end = true
}

func (r *restart) isDone() bool {
	return r.end
}

func restartRun(cmd *cobra.Command, args []string) {
	project, container, err := cmdcontext.GetProjectOrContainerID(args)

	if err != nil {
		if err = cmd.Help(); err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	var r = &restart{
		project:   project,
		container: container,
	}

	if quiet {
		r.do()
		return
	}

	var queue sync.WaitGroup

	queue.Add(2)

	go func() {
		r.do()
		queue.Done()
	}()

	go func() {
		r.watch()
		queue.Done()
	}()

	queue.Wait()
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
