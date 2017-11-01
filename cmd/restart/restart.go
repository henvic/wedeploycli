package restart

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
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

var setupHost = cmdflagsfromhost.SetupHost{
	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
	},
	Pattern: cmdflagsfromhost.FullHostPattern,
}

func init() {
	setupHost.Init(RestartCmd)

	// the --quiet parameter was removed
	_ = RestartCmd.Flags().BoolP("quiet", "q", false, "")
	_ = RestartCmd.Flags().MarkHidden("quiet")
}

type restart struct {
	project string
	service string
	ctx     context.Context
}

func (r *restart) do() (err error) {
	switch r.service {
	case "":
		projectsClient := projects.New(we.Context())
		err = projectsClient.Restart(r.ctx, r.project)
	default:
		servicesClient := services.New(we.Context())
		err = servicesClient.Restart(r.ctx, r.project, r.service)
	}

	if err != nil {
		return err
	}

	switch r.service {
	case "":
		fmt.Printf("Restarting project %s.\n", r.project)
	default:
		fmt.Printf("Restarting service %s on project %s.\n", r.service, r.project)
	}

	return nil
}

func (r *restart) checkProjectOrServiceExists() (err error) {
	if r.service == "" {
		projectsClient := projects.New(we.Context())
		_, err = projectsClient.Get(r.ctx, r.project)
	} else {
		servicesClient := services.New(we.Context())
		_, err = servicesClient.Get(r.ctx, r.project, r.service)
	}

	return err
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(we.Context())
}

func restartRun(cmd *cobra.Command, args []string) error {
	var r = &restart{
		project: setupHost.Project(),
		service: setupHost.Service(),
	}

	r.ctx = context.Background()

	if err := r.checkProjectOrServiceExists(); err != nil {
		return err
	}

	return r.do()
}
