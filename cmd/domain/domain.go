package cmddomain

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	cmddomainadd "github.com/wedeploy/cli/cmd/domain/add"
	cmddomainremove "github.com/wedeploy/cli/cmd/domain/remove"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/services"
)

// DomainCmd controls the domains for a given project
var DomainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Show and configure domain names for services",
	Long: `Show and configure domain names for services

Changing these values does not change wedeploy.json hard coded values.

Information about name servers configuration at
http://wedeploy.com/docs/intro/custom-domains.html`,
	Example: `  we domain (to list domains)
  we domain add foo.com
  we domain rm foo.com`,
	PreRunE: preRun,
	RunE:    run,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern:             cmdflagsfromhost.FullHostPattern,
	UseProjectDirectory: true,
	UseServiceDirectory: true,
	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
		Service: true,
	},
}

func preRun(cmd *cobra.Command, args []string) error {
	// get nice error message
	if _, _, err := cmd.Find(args); err != nil {
		return err
	}

	return setupHost.Process()
}

func run(cmd *cobra.Command, args []string) error {
	var service, err = services.Get(context.Background(),
		setupHost.Project(),
		setupHost.Service())

	if err != nil {
		return err
	}

	for _, customDomain := range service.CustomDomains {
		fmt.Println(customDomain)
	}

	return nil
}

func init() {
	setupHost.Init(DomainCmd)
	DomainCmd.AddCommand(cmddomainadd.Cmd)
	DomainCmd.AddCommand(cmddomainremove.Cmd)
}
