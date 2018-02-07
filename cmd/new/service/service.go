package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/wedeploy/cli/fancy"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/services"
)

var image string

var setupHost = cmdflagsfromhost.SetupHost{
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
	RunE: (&newService{}).run,
	Args: cobra.NoArgs,
}

type newService struct {
	servicesClient *services.Client
}

func (ns *newService) getOrSelectImageType() (string, error) {
	if image != "" {
		return image, nil
	}

	var catalog, err = ns.servicesClient.Catalog(context.Background())

	var options = fancy.Options{
		Hash: true,
	}

	for n, i := range catalog {
		options.Add(fmt.Sprintf("%d", n+1), "\t"+i.Name+"    \t"+color.Format(color.FgHiBlack, i.Image))
	}

	option, err := options.Ask("Select a Service Type")

	if err != nil {
		return "", err
	}

	choice, err := strconv.Atoi(option)

	switch {
	case err == nil:
		return catalog[choice-1].Image, nil
	case choice < 1 || choice > len(catalog):
		return "", errors.New("invalid option")
	default:
		return option, nil
	}
}

func (ns *newService) run(cmd *cobra.Command, args []string) error {
	ns.servicesClient = services.New(we.Context())

	if err := setupHost.Process(context.Background(), we.Context()); err != nil {
		return err
	}

	var imageType, err = ns.getOrSelectImageType()

	if err != nil {
		return err
	}

	var serviceID = setupHost.Service()

	if serviceID == "" {
		fmt.Println(fancy.Question("Choose a Service ID") + " " + fancy.Tip("default: random"))
		serviceID, err = fancy.Prompt()

		if err != nil {
			return err
		}
	}

	body := services.CreateBody{
		ServiceID: serviceID,
		Image:     imageType,
	}

	s, err := ns.servicesClient.Create(context.Background(), setupHost.Project(), body)

	if err != nil {
		return err
	}

	fmt.Printf(color.Format(color.FgHiBlack, "Service \"")+
		"%s-%s.%s"+
		color.Format(color.FgHiBlack, "\" created.")+"\n",
		s.ServiceID,
		setupHost.Project(),
		setupHost.InfrastructureDomain())

	return nil
}

func init() {
	Cmd.Flags().StringVar(&image, "image", "", "Image type")

	setupHost.Init(Cmd)
}
