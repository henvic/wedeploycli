package root

import (
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/command/about"
	"github.com/wedeploy/cli/command/activities"
	"github.com/wedeploy/cli/command/autocomplete"
	"github.com/wedeploy/cli/command/console"
	"github.com/wedeploy/cli/command/curl"
	"github.com/wedeploy/cli/command/delete"
	"github.com/wedeploy/cli/command/deploy"
	"github.com/wedeploy/cli/command/diagnostics"
	"github.com/wedeploy/cli/command/docs"
	"github.com/wedeploy/cli/command/domain"
	"github.com/wedeploy/cli/command/env-var"
	"github.com/wedeploy/cli/command/exec"
	"github.com/wedeploy/cli/command/inspect"
	"github.com/wedeploy/cli/command/list"
	"github.com/wedeploy/cli/command/log"
	"github.com/wedeploy/cli/command/login"
	"github.com/wedeploy/cli/command/logout"
	"github.com/wedeploy/cli/command/metrics"
	"github.com/wedeploy/cli/command/new"
	"github.com/wedeploy/cli/command/open"
	"github.com/wedeploy/cli/command/remote"
	"github.com/wedeploy/cli/command/restart"
	"github.com/wedeploy/cli/command/scale"
	"github.com/wedeploy/cli/command/shell"
	"github.com/wedeploy/cli/command/uninstall"
	"github.com/wedeploy/cli/command/update"
	versioncmd "github.com/wedeploy/cli/command/version"
	"github.com/wedeploy/cli/command/who"
)

var commands = []*cobra.Command{
	activities.ActivitiesCmd,
	deploy.DeployCmd,
	list.ListCmd,
	new.NewCmd,
	log.LogCmd,
	domain.DomainCmd,
	env.EnvCmd,
	scale.ScaleCmd,
	restart.RestartCmd,
	delete.DeleteCmd,
	exec.ExecCmd,
	shell.ShellCmd,
	login.LoginCmd,
	logout.LogoutCmd,
	open.OpenCmd,
	console.ConsoleCmd,
	docs.DocsCmd,
	autocomplete.AutocompleteCmd,
	remote.RemoteCmd,
	diagnostics.DiagnosticsCmd,
	update.UpdateCmd,
	uninstall.UninstallCmd,
	versioncmd.VersionCmd,
	inspect.InspectCmd,
	who.WhoCmd,
	about.AboutCmd,
	metrics.MetricsCmd,
	curl.CurlCmd,
}
