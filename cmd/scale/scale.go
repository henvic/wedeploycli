package scale

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/isterm"
	"github.com/wedeploy/cli/services"
)

// ScaleCmd is used for getting scale
var ScaleCmd = &cobra.Command{
	Use:   "scale",
	Short: "Configure number of instances for services",
	RunE:  scaleRun,
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

func getInstancesNumber(cmd *cobra.Command, args []string) (string, error) {
	if len(args) != 0 || !isterm.Check() {
		err := cobra.ExactArgs(1)(cmd, args)
		return args[0], err
	}

	fmt.Println(fancy.Question("Number of instances"))

	var sscale, err = fancy.Prompt()

	if err != nil {
		return "", err
	}

	return sscale, nil
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

func isValidNumberOfInstances(instances string) bool {
	n, err := strconv.Atoi(instances)
	return err == nil && n > 0
}

func scaleRun(cmd *cobra.Command, args []string) error {
	sscale, err := getInstancesNumber(cmd, args)

	if err != nil {
		return err
	}

	if !isValidNumberOfInstances(sscale) {
		return fmt.Errorf(`"%v" isn't a valid number for instances`, sscale)
	}

	if err := setupHost.Process(context.Background(), we.Context()); err != nil {
		return err
	}

	current, err := strconv.Atoi(sscale)

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
