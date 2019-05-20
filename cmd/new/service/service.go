package service

import (
	"context"
	"fmt"

	"github.com/wedeploy/cli/fancy"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/services"
)

var image string

// Don't use this anywhere but on Cmd.RunE
var sh = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Project: true,
		Auth:    true,
	},

	PromptMissingProject: true,
	HideServicesPrompt:   true,
}

// Cmd is the command for installing a new service
var Cmd = &cobra.Command{
	Use:   "service",
	Short: "Install new service",
	// Don't use other run functions besides RunE here
	// or fix NewCmd to call it correctly
	RunE: runE,
	Args: cobra.NoArgs,
}

func runE(cmd *cobra.Command, args []string) error {
	if err := sh.Process(context.Background(), we.Context()); err != nil {
		return err
	}

	return (&newService{}).run(sh.Project(), sh.Service(), sh.ServiceDomain())
}

// Run command for creating a service
func Run(projectID, serviceID, serviceDomain string) error {
	return (&newService{}).run(projectID, serviceID, serviceDomain)
}

type newService struct {
	servicesClient *services.Client
}

func (n *newService) getOrSelectImageType() (string, error) {
	if image != "" {
		return image, nil
	}

	fmt.Println(fancy.Question("Type a service type"))
	return fancy.Prompt()
}

func (n *newService) run(projectID, serviceID, serviceDomain string) error {
	wectx := we.Context()
	n.servicesClient = services.New(wectx)

	var err error

	if serviceID == "" {
		fmt.Println(fancy.Question("Choose a Service ID") + " " + fancy.Tip("default: random"))
		serviceID, err = fancy.Prompt()

		if err != nil {
			return err
		}
	}

	fmt.Println("")

	imageType, err := n.getOrSelectImageType()

	if err != nil {
		return err
	}

	body := services.CreateBody{
		ServiceID: serviceID,
		Image:     imageType,
	}

	s, err := n.servicesClient.Create(context.Background(), projectID, body)

	if err != nil {
		return err
	}

	fmt.Printf(color.Format(color.FgHiBlack, "Service \"")+
		"%s-%s.%s"+
		color.Format(color.FgHiBlack, "\" created on ")+
		wectx.InfrastructureDomain()+
		color.Format(color.FgHiBlack, ".")+
		"\n",
		s.ServiceID,
		projectID,
		serviceDomain)

	return nil
}

func init() {
	Cmd.Flags().StringVar(&image, "image", "", "Image type")

	sh.Init(Cmd)
}
