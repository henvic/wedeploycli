// Hack to make git push work fine on Windows
// without a git-credential-helper issue / Invalid credentials error.
// See https://github.com/wedeploy/cli/issues/323
// This passes the token as part of the remote address
// Security risk: prone to sniffing (on the same machine) on most operating systems.

package deployment

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	version "github.com/hashicorp/go-version"
	"github.com/wedeploy/cli/envs"
	"github.com/wedeploy/cli/verbose"
)

// Tests were made on the following git versions on Windows:
// 2.5.0, 2.5.1, 2.5.2, 2.5.2.2, 2.5.3, 2.6.0, 2.6.4, 2.7.4, 2.8.4,
// 2.9.0, 2.9.2, 2.10.1, 2.12.1, 2.13.0, 2.13.3, 2.14.2.2, 2.14.2.3.
// This issue doesn't appear to affect git on Linux (not even with 1.9.1).
// Related: https://github.com/git-for-windows/git

// Special treatment constraints:
// 2.5.0 (Aug 18 2015): weird git credential- error message, but still works
// 2.5.1 (Aug 28 2015): weird git credential- error message, but still works
// 2.5.3 (Sep 18 2017): starts breaking
// 2.13.3 (Jul 13 2017): working again
const gitAffectedVersions = "> 2.5.1, < 2.13.3"

func (d *Deploy) pushHack() (groupUID string, err error) {
	d.pushStartTime = time.Now()
	defer func() {
		d.pushEndTime = time.Now()
	}()

	var params = []string{"push", d.getGitRemote(), "master", "--force"}

	if verbose.Enabled {
		params = append(params, "--verbose")
	}

	verbose.Debug(fmt.Sprintf("Running git push %v master -force",
		verbose.SafeEscape(d.getGitRemote())))

	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(d.getConfigEnvs(),
		"GIT_TERMINAL_PROMPT=0",
		envs.GitCredentialRemoteToken+"="+d.ConfigContext.Token(),
	)
	cmd.Dir = d.Path

	var bufErr *bytes.Buffer

	switch verbose.IsUnsafeMode() {
	case true:
		bufErr = copyErrStreamAndVerbose(cmd)
	default:
		bufErr = &bytes.Buffer{}
		cmd.Stderr = bufErr
	}

	err = cmd.Run()

	if err != nil {
		bs := bufErr.String()
		switch {
		// I need to see if there are any "error:" strings as well
		case strings.Contains(bs, "fatal: Authentication failed for"),
			strings.Contains(bs, "could not read Username"):
			return "", errors.New("invalid credentials: please update git and try again http://git-scm.com")
		case strings.Contains(bs, "error: "):
			return "", fmt.Errorf("git error: %v", verbose.SafeEscape(getGitErrors(bs).Error()))
		default:
			return "", err
		}
	}

	return tryGetPushGroupUID(*bufErr)
}

func (d *Deploy) addRemoteHack() error {
	verbose.Debug("Adding remote with token")
	var gitServer = fmt.Sprintf("https://%v:@git.%v/%v.git",
		d.ConfigContext.Token(),
		d.ConfigContext.InfrastructureDomain(),
		d.ProjectID)

	var params = []string{"remote", "add", d.getGitRemote(), gitServer}

	verbose.Debug(fmt.Sprintf("Running git remote add %v %v",
		d.getGitRemote(),
		verbose.SafeEscape(gitServer)))

	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path

	if verbose.IsUnsafeMode() {
		cmd.Stderr = errStream
	}

	return cmd.Run()
}

func (d *Deploy) useGitCredentialHack() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	v, err := version.NewVersion(d.gitVersion)

	if err != nil {
		return false
	}

	constraints, err := version.NewConstraint(gitAffectedVersions)

	if err != nil {
		verbose.Debug(err)
		return false
	}

	p := constraints.Check(v)

	if p {
		verbose.Debug("git version " + d.gitVersion + " is not compatible with credential-helper due to a bug")
		verbose.Debug("fall back to passing token on remote")
		verbose.Debug("limited debug messages for security reasons")
		verbose.Debug("updating git is highly recommended")
	}

	return p
}
