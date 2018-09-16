package feedback

import (
	"fmt"
	"os"

	"github.com/henvic/browser"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/errorhandler"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/verbose"
)

func (w *Watch) maybeSetOpenLogsFunc() {
	var f = w.f

	if len(f.BuildFailed) == 0 && len(f.DeployFailed) == 0 {
		return
	}

	errorhandler.SetAfterError(func() {
		w.maybeOpenLogs()
	})
}

func (w *Watch) maybeOpenLogs() {
	var f = w.f
	var fb = f.BuildFailed
	var fd = f.DeployFailed

	switch y, err := fancy.Boolean("Open browser to check the logs?"); {
	case err != nil:
		_, _ = fmt.Fprintf(os.Stderr, "%v", err)
		fallthrough
	case !y:
		return
	}

	var u = fmt.Sprintf("https://%v%v/projects/%v/logs",
		defaults.DashboardAddressPrefix,
		w.ConfigContext.InfrastructureDomain(),
		w.ProjectID)

	switch {
	case len(fb) == 1 && len(fd) == 0:
		u += "?label=buildUid&logServiceId=" + fb[0]
	case len(fb) == 0 && len(fd) == 1:
		u += "?logServiceId=" + fd[0]
	case len(fd) == 0:
		u += "?label=buildUid"
	}

	if err := browser.OpenURL(u); err != nil {
		fmt.Println("Open URL: (can't open automatically)", u)
		verbose.Debug(err)
	}
}
