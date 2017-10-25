package root

import (
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/about"
	"github.com/wedeploy/cli/cmd/activities"
	"github.com/wedeploy/cli/cmd/autocomplete"
	"github.com/wedeploy/cli/cmd/console"
	"github.com/wedeploy/cli/cmd/delete"
	"github.com/wedeploy/cli/cmd/deploy"
	"github.com/wedeploy/cli/cmd/diagnostics"
	"github.com/wedeploy/cli/cmd/docs"
	"github.com/wedeploy/cli/cmd/domain"
	"github.com/wedeploy/cli/cmd/env"
	"github.com/wedeploy/cli/cmd/gitcredentialhelper"
	"github.com/wedeploy/cli/cmd/inspect"
	"github.com/wedeploy/cli/cmd/list"
	"github.com/wedeploy/cli/cmd/log"
	"github.com/wedeploy/cli/cmd/login"
	"github.com/wedeploy/cli/cmd/logout"
	"github.com/wedeploy/cli/cmd/remote"
	"github.com/wedeploy/cli/cmd/restart"
	"github.com/wedeploy/cli/cmd/uninstall"
	"github.com/wedeploy/cli/cmd/update"
	versioncmd "github.com/wedeploy/cli/cmd/version"
	"github.com/wedeploy/cli/cmd/who"
)

var commands = []*cobra.Command{
	activities.ActivitiesCmd,
	deploy.DeployCmd,
	list.ListCmd,
	console.ConsoleCmd,
	docs.DocsCmd,
	log.LogCmd,
	domain.DomainCmd,
	env.EnvCmd,
	restart.RestartCmd,
	delete.DeleteCmd,
	login.LoginCmd,
	logout.LogoutCmd,
	autocomplete.AutocompleteCmd,
	remote.RemoteCmd,
	diagnostics.DiagnosticsCmd,
	versioncmd.VersionCmd,
	update.UpdateCmd,
	inspect.InspectCmd,
	who.WhoCmd,
	gitcredentialhelper.GitCredentialHelperCmd,
	uninstall.UninstallCmd,
	about.AboutCmd,
}