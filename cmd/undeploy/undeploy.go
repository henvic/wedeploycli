package cmdundeploy

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
)

// UndeployCmd undeploys a given project or container
var UndeployCmd = &cobra.Command{
	Use:     "undeploy",
	Short:   "Undeploy a given project or container",
	PreRunE: preRun,
	RunE:    run,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern:               cmdflagsfromhost.FullHostPattern,
	UseProjectDirectory:   true,
	UseContainerDirectory: true,
	Requires: cmdflagsfromhost.Requires{
		Auth: true,
	},
}

func init() {
	setupHost.Init(UndeployCmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process()
}

func run(cmd *cobra.Command, args []string) error {
	if setupHost.Remote() == "" {
		return errors.New(`You can not run undeploy in the local infrastructure. Use "we dev stop" instead`)
	}

	var project = setupHost.Project()
	var container = setupHost.Container()

	if container != "" {
		return containers.Unlink(context.Background(), project, container)
	}

	return projects.Unlink(context.Background(), project)
}
