package cmddomainremove

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/projects"
)

// Cmd for removing a domain
var Cmd = &cobra.Command{
	Use:     "remove",
	Short:   "Remove custom domain of a given project",
	Example: "we domain remove --domain example.com",
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
	if err := cmdargslen.Validate(args, 0, 0); err != nil {
		return err
	}

	return setupHost.Process()
}

func run(cmd *cobra.Command, args []string) error {
	if domain == "" {
		return errors.New("Custom domain is required")
	}

	return projects.RemoveDomain(
		context.Background(),
		setupHost.Project(),
		domain)
}
