package cmddomainadd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/projects"
)

// Cmd for adding a domain
var Cmd = &cobra.Command{
	Use:     "add",
	Short:   "Add custom domain to a given project",
	Example: "we domain add example.com",
	PreRunE: preRun,
	RunE:    run,
}

var domain string

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern:             cmdflagsfromhost.ProjectAndRemotePattern,
	UseProjectDirectory: true,
	Requires: cmdflagsfromhost.Requires{
		Auth:    true,
		Project: true,
	},
}

func init() {
	setupHost.Init(Cmd)
	Cmd.Flags().StringVar(&domain, "domain", "",
		"Custom domain")
}

func preRun(cmd *cobra.Command, args []string) error {
	if err := cmdargslen.Validate(args, 1, 1); err != nil {
		return err
	}

	return setupHost.Process()
}

func run(cmd *cobra.Command, args []string) error {
	return projects.AddDomain(
		context.Background(),
		setupHost.Project(),
		args[0])
}
