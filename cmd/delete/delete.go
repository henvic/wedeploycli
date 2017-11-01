package delete

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/canceled"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdargslen"
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
	Hidden:  true,
	Short:   "Delete project or services",
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
	if len(args) != 0 {
		return errors.New("invalid number of arguments")
	}

	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process(we.Context())
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

	if err := confirmation(); err != nil {
		return err
	}

	return u.do()
}

func confirmation() error {
	if force {
		return nil
	}

	var question string

	if setupHost.Service() != "" {
		question = fmt.Sprintf(`Do you really want to delete the service "%v" on project "%v"?`,
			color.Format(color.Bold, setupHost.Service()),
			color.Format(color.Bold, setupHost.Project()))
	} else {
		question = fmt.Sprintf(`Do you really want to delete the project "%v"?`,
			color.Format(color.Bold, setupHost.Project()))
	}

	var confirm, askErr = fancy.Boolean(question)

	if askErr != nil {
		return askErr
	}

	if confirm {
		return nil
	}

	return canceled.CancelCommand("delete canceled")
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
		fmt.Printf("Deleting project %s.\n", u.project)
	default:
		fmt.Printf("Deleting service %s on project %s.\n", u.service, u.project)
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
