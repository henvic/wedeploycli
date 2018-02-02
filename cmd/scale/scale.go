package scale

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/services"
)

// ScaleCmd is used for getting scale
var ScaleCmd = &cobra.Command{
	Use:     "scale",
	Short:   "Configure number of instances for services",
	Args:    cobra.ExactArgs(1),
	PreRunE: preRun,
	RunE:    scaleRun,
	Example: `  we scale --project chat --service data 3
  we scale --project chat --service data --remote wedeploy 5
  we scale --url data-chat.wedeploy.io 1`,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
		Service: true,
	},

	PromptMissingProject: true,
	PromptMissingService: true,
}

func init() {
	setupHost.Init(ScaleCmd)
}

type scale struct {
	ctx context.Context

	project string
	service string
	current int
}

func (s *scale) do() (err error) {
	servicesClient := services.New(we.Context())
	err = servicesClient.Scale(s.ctx, s.project, s.service, services.Scale{
		Current: s.current,
	})

	if err == nil {
		fmt.Printf("Setting the number of instances for \"%s\" to %d.\n", setupHost.Host(), s.current)
	}

	return err
}

func (s *scale) checkServiceExists() (err error) {
	servicesClient := services.New(we.Context())
	_, err = servicesClient.Get(s.ctx, s.project, s.service)
	return err
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func scaleRun(cmd *cobra.Command, args []string) error {
	current, err := strconv.Atoi(args[0])

	if err != nil {
		return err
	}

	var s = &scale{
		ctx: context.Background(),

		project: setupHost.Project(),
		service: setupHost.Service(),
		current: current,
	}

	if err := s.checkServiceExists(); err != nil {
		return err
	}

	return s.do()
}
