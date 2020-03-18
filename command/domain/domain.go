package domain

import (
	"context"

	"github.com/henvic/wedeploycli/cmdflagsfromhost"
	cmddomainadd "github.com/henvic/wedeploycli/command/domain/add"
	"github.com/henvic/wedeploycli/command/domain/internal/commands"
	cmddomainremove "github.com/henvic/wedeploycli/command/domain/remove"
	cmddomainshow "github.com/henvic/wedeploycli/command/domain/show"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/fancy"
	"github.com/henvic/wedeploycli/services"
	"github.com/spf13/cobra"
)

type interativeDomainCmd struct {
	ctx context.Context
	c   commands.Command
}

var idc = &interativeDomainCmd{}

// DomainCmd controls the domains for a given project
var DomainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Show and configure domain names for services",
	Example: `  lcp domain (to list domains)
  lcp domain add foo.com
  lcp domain rm foo.com`,
	Args:    cobra.NoArgs,
	PreRunE: idc.preRun,
	RunE:    idc.run,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.FullHostPattern,

	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
		Service: true,
	},

	PromptMissingService: true,
}

func (idc *interativeDomainCmd) preRun(cmd *cobra.Command, args []string) error {
	idc.ctx = context.Background()

	// get nice error message
	if _, _, err := cmd.Find(args); err != nil {
		return err
	}

	return setupHost.Process(context.Background(), we.Context())
}

func (idc *interativeDomainCmd) run(cmd *cobra.Command, args []string) error {
	idc.c = commands.Command{
		SetupHost:      setupHost,
		ServicesClient: services.New(we.Context()),
	}

	if err := idc.c.Show(idc.ctx); err != nil {
		return err
	}

	var operations = fancy.Options{}

	const (
		addOption   = "a"
		unsetOption = "d"
	)

	operations.Add(addOption, "Add custom domain")
	operations.Add(unsetOption, "Delete custom domain")

	op, err := operations.Ask("Select one of the operations for \"" + setupHost.Host() + "\":")

	if err != nil {
		return err
	}

	switch op {
	case unsetOption:
		err = idc.deleteCmd()
	default:
		err = idc.addCmd()
	}

	if err != nil {
		return err
	}

	return idc.c.Show(idc.ctx)
}

func (idc *interativeDomainCmd) addCmd() error {
	return idc.c.Add(idc.ctx, []string{})
}

func (idc *interativeDomainCmd) deleteCmd() error {
	return idc.c.Delete(idc.ctx, []string{})
}

func init() {
	setupHost.Init(DomainCmd)
	DomainCmd.AddCommand(cmddomainshow.Cmd)
	DomainCmd.AddCommand(cmddomainadd.Cmd)
	DomainCmd.AddCommand(cmddomainremove.Cmd)
}
