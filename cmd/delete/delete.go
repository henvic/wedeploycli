package delete

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/canceled"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
)

var force bool

// DeleteCmd is the delete command to undeploy a project or service
var DeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete project or services\n\t\t",
	Args:    cobra.NoArgs,
	PreRunE: preRun,
	RunE:    run,
	Example: `  we delete --url data.chat.wedeploy.io
  we delete --project chat
  we delete --project chat --service data`,
}

type undeployer struct {
	context              context.Context
	project              string
	service              string
	infrastructureDomain string
	serviceDomain        string
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Project: true,
		Auth:    true,
	},

	PromptMissingProject: true,
}

func init() {
	// the --quiet parameter was removed
	_ = DeleteCmd.Flags().BoolP("quiet", "q", false, "")
	_ = DeleteCmd.Flags().MarkHidden("quiet")

	DeleteCmd.Flags().BoolVar(&force, "force", false,
		"Force deleting services without confirmation")
	setupHost.Init(DeleteCmd)
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func run(cmd *cobra.Command, args []string) error {
	var u = undeployer{
		context:              context.Background(),
		project:              setupHost.Project(),
		service:              setupHost.Service(),
		infrastructureDomain: setupHost.InfrastructureDomain(),
		serviceDomain:        setupHost.ServiceDomain(),
	}

	if err := u.checkProjectOrServiceExists(); err != nil {
		return err
	}

	if err := u.confirmation(); err != nil {
		return err
	}

	return u.do()
}

func (u *undeployer) confirmation() error {
	if force {
		return nil
	}

	if u.service != "" {
		fmt.Print(color.Format(color.FgRed, color.Bold, "Deleting a service cannot be undone."))
		fmt.Println(" All the access and domains related to this service will be lost.")
		fmt.Printf("Enter the service ID %s to delete it permanently:\n",
			color.Format(color.ReverseVideo, u.service))
	} else {
		fmt.Print(color.Format(color.FgRed, color.Bold, "Deleting a project cannot be undone."))
		fmt.Println(" All the services and the access and domains related to this project will be lost.")
		fmt.Printf("Enter the project ID %s to delete it permanently:\n",
			color.Format(color.ReverseVideo, u.project))
	}

	var verify, askErr = fancy.Prompt()

	if askErr != nil {
		return askErr
	}

	switch {
	case verify == "":
		return canceled.Skip()
	case (u.service != "" && verify == u.service) || u.service == "" && verify == u.project:
		return nil
	default:
		return errors.New("invalid value")
	}
}

func (u *undeployer) do() (err error) {
	switch u.service {
	case "":
		projectsClient := projects.New(we.Context())
		err = projectsClient.Unlink(u.context, u.project)
	default:
		servicesClient := services.New(we.Context())
		err = servicesClient.Unlink(u.context, u.project, u.service)
	}

	if err != nil {
		return err
	}

	switch u.service {
	case "":
		fmt.Printf(color.Format(
			color.FgHiBlack, "Deleting project \"")+
			"%s"+
			color.Format(color.FgHiBlack, "\".")+
			"\n",
			u.project)
	default:
		fmt.Printf(color.Format(
			color.FgHiBlack, "Deleting service \"")+
			"%v"+
			color.Format(color.FgHiBlack, "\" on project \"")+
			"%v"+
			color.Format(color.FgHiBlack, "\".")+
			"\n",
			u.service,
			u.project)
	}

	return nil
}

func (u *undeployer) checkProjectOrServiceExists() (err error) {
	if u.service == "" {
		projectsClient := projects.New(we.Context())
		_, err = projectsClient.Get(u.context, u.project)
	} else {
		servicesClient := services.New(we.Context())
		_, err = servicesClient.Get(u.context, u.project, u.service)
	}

	return err
}
