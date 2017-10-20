// +build windows

package deployment

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/envs"
	"github.com/wedeploy/cli/metrics"
	"github.com/wedeploy/cli/verbose"
)

// Push deployment to the WeDeploy remote
func (d *Deploy) Push() (groupUID string, err error) {
	d.pushStartTime = time.Now()
	defer func() {
		d.pushEndTime = time.Now()
	}()

	var params = []string{"push", d.getGitRemote(), "master", "--force"}

	if verbose.Enabled {
		params = append(params, "--verbose")
	}

	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(d.getConfigEnvs(),
		"GIT_TERMINAL_PROMPT=0",
		envs.GitCredentialRemoteToken+"="+d.ConfigContext.Token(),
	)
	cmd.Dir = d.Path

	var bufErr = copyErrStreamAndVerbose(cmd)
	err = cmd.Run()

	if err != nil {
		bs := bufErr.String()
		switch {
		// I need to see if there are any "error:" strings as well
		case strings.Contains(bs, "fatal: Authentication failed for"),
			strings.Contains(bs, "could not read Username"):
			metrics.Rec(s.wectx.Config(), metrics.Event{
				Type: "push_on_windows_invalid_credentials_failure",
				Text: "Invalid credentials failure on Windows, trying again.",
			})
			return d.pushAfterInvalidCredentialsFailure()
		case strings.Contains(bs, "error: "):
			return "", getGitErrors(bs)
		default:
			return "", err
		}
	}

	return tryGetPushGroupUID(*bufErr)
}

// Push deployment to the WeDeploy remote
func (d *Deploy) pushAfterInvalidCredentialsFailure() (groupUID string, err error) {
	d.pushStartTime = time.Now()
	defer func() {
		d.pushEndTime = time.Now()
	}()

	var params = []string{"push", d.getGitRemote(), "master", "--force"}

	if verbose.Enabled {
		params = append(params, "--verbose")
	}

	verbose.Debug(fmt.Sprintf("Running git %v (with authentication by extra-header)",
		strings.Join(params, " ")))

	verbose.Debug(color.Format(color.BgYellow, "Basic Auth credential: hidden value"))

	var tokenHeader = fmt.Sprintf("http.%s.extraHeader=Authorization: %s",
		d.GitRemoteAddress,
		d.ConfigContext.Token())

	params = append(
		[]string{"-c", tokenHeader},
		params...)

	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(d.getConfigEnvs(),
		"GIT_TERMINAL_PROMPT=0",
		envs.GitCredentialRemoteToken+"="+d.ConfigContext.Token(),
	)
	cmd.Dir = d.Path

	verbose.Debug("Hidding stderr from second git push try on Windows for safety reasons")
	var bufErr bytes.Buffer
	cmd.Stderr = &bufErr
	err = cmd.Run()

	if err != nil {
		verbose.Debug("Second git push try stderr:")
		bs := bufErr.String()
		verbose.SafeEscape(bs)
		return "", errors.New("Invalid credentials (second try)")
	}

	return tryGetPushGroupUID(*bufErr)
}
