package cmdrestart

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/verbose"
)

// RestartCmd is used for getting restart
var RestartCmd = &cobra.Command{
	Use:     "restart --url <host URL> or --project <project> --container <container>",
	Short:   "Restart project or container\n",
	PreRunE: preRun,
	RunE:    restartRun,
	Example: `we restart --project chat --container data
we restart --container data
we restart --project chat --container data
we restart --project chat --container data --remote cloud
we restart --url data.chat.wedeploy.me`,
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
	err       error
	end       bool
}

func (r *restart) do() {
	switch r.container {
	case "":
		r.err = projects.Restart(context.Background(), r.project)
	default:
		r.err = containers.Restart(context.Background(), r.project, r.container)
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

	var p = r.list.Projects[0]

	if p.Health != "up" {
		return false
	}

	if r.container == "" {
		return true
	}

	var cs, ec = p.Services(context.Background())

	if ec != nil {
		fmt.Fprintf(os.Stderr, "Can't check if containers are finished: %v\n", ec)
		return false
	}

	for _, c := range cs {
		if c.Health != "up" {
			return false
		}
	}

	return true
}

func (r *restart) checkProjectOrContainerExists() error {
	var err error
	if r.container == "" {
		_, err = projects.Get(context.Background(), r.project)
	} else {
		_, err = containers.Get(context.Background(), r.project, r.container)
	}

	return err
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process()
}

func restartRun(cmd *cobra.Command, args []string) error {
	var r = &restart{
		project:   setupHost.Project(),
		container: setupHost.Container(),
	}

	if err := r.checkProjectOrContainerExists(); err != nil {
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
