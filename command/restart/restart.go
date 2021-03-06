package restart

import (
	"context"
	"fmt"

	"github.com/henvic/wedeploycli/cmdflagsfromhost"
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/services"
	"github.com/spf13/cobra"
)

// RestartCmd is used for getting restart
var RestartCmd = &cobra.Command{
	Use:     "restart",
	Short:   "Restart services",
	Args:    cobra.NoArgs,
	PreRunE: preRun,
	RunE:    restartRun,
	Example: `  lcp restart --project chat --service data
^  lcp restart --project chat --service data --remote lfr-cloud
^  lcp restart --url data-chat.lfr.cloud`,
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
	wectx := we.Context()
	servicesClient := services.New(wectx)
	err = servicesClient.Restart(r.ctx, r.project, r.service)

	if err == nil {
		fmt.Printf(color.Format(color.FgHiBlack,
			"Restarting service \"")+
			"%s"+color.Format(color.FgHiBlack,
			"\" on project \"")+
			"%s"+
			color.Format(color.FgHiBlack, "\" on ")+
			wectx.InfrastructureDomain()+
			color.Format(color.FgHiBlack, ".")+
			"\n",
			r.service,
			r.project)
	}

	return err
}

func (r *restart) checkServiceExists() (err error) {
	servicesClient := services.New(we.Context())
	_, err = servicesClient.Get(r.ctx, r.project, r.service)
	return err
}

func preRun(cmd *cobra.Command, args []string) error {
	return setupHost.Process(context.Background(), we.Context())
}

func restartRun(cmd *cobra.Command, args []string) error {
	var r = &restart{
		project: setupHost.Project(),
		service: setupHost.Service(),
		ctx:     context.Background(),
	}

	if err := r.checkServiceExists(); err != nil {
		return err
	}

	return r.do()
}
