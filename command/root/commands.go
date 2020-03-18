package root

import (
	"github.com/henvic/wedeploycli/command/about"
	"github.com/henvic/wedeploycli/command/activities"
	"github.com/henvic/wedeploycli/command/autocomplete"
	"github.com/henvic/wedeploycli/command/console"
	"github.com/henvic/wedeploycli/command/curl"
	"github.com/henvic/wedeploycli/command/delete"
	"github.com/henvic/wedeploycli/command/deploy"
	"github.com/henvic/wedeploycli/command/diagnostics"
	"github.com/henvic/wedeploycli/command/docs"
	"github.com/henvic/wedeploycli/command/domain"
	"github.com/henvic/wedeploycli/command/env-var"
	"github.com/henvic/wedeploycli/command/exec"
	"github.com/henvic/wedeploycli/command/inspect"
	"github.com/henvic/wedeploycli/command/list"
	"github.com/henvic/wedeploycli/command/log"
	"github.com/henvic/wedeploycli/command/login"
	"github.com/henvic/wedeploycli/command/logout"
	"github.com/henvic/wedeploycli/command/metrics"
	"github.com/henvic/wedeploycli/command/new"
	"github.com/henvic/wedeploycli/command/open"
	"github.com/henvic/wedeploycli/command/remote"
	"github.com/henvic/wedeploycli/command/restart"
	"github.com/henvic/wedeploycli/command/scale"
	"github.com/henvic/wedeploycli/command/shell"
	"github.com/henvic/wedeploycli/command/uninstall"
	"github.com/henvic/wedeploycli/command/update"
	versioncmd "github.com/henvic/wedeploycli/command/version"
	"github.com/henvic/wedeploycli/command/who"
	"github.com/spf13/cobra"
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
