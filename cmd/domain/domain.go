package cmddomain

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	cmddomainadd "github.com/wedeploy/cli/cmd/domain/add"
	cmddomainremove "github.com/wedeploy/cli/cmd/domain/remove"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/projects"
)

// DomainCmd controls the domains for a given project
var DomainCmd = &cobra.Command{
	Use:     "domain",
	Aliases: []string{"set"},
	Short:   "Configure custom domains for projects",
	Long: `Configure custom domains for projects

use "we domain" to list domains on the infrastructure.

Changing these values does not change container.json hard coded values.

Information about name servers configuration at
http://wedeploy.com/docs/intro/custom-domains.html`,
	Example: `  we domain
  we domain add foo.com
  we domain rm foo.com`,
	PreRunE: preRun,
	RunE:    run,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern:             cmdflagsfromhost.ProjectAndRemotePattern,
	UseProjectDirectory: true,
	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
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
	var project, err = projects.Get(context.Background(), setupHost.Project())

	if err != nil {
		return err
	}

	for _, customDomain := range project.CustomDomains {
		fmt.Println(customDomain)
	}

	return nil
}

func init() {
	setupHost.Init(DomainCmd)
	DomainCmd.AddCommand(cmddomainadd.Cmd)
	DomainCmd.AddCommand(cmddomainremove.Cmd)
}