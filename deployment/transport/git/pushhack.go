// Hack to make git push work fine on Windows
// without a git-credential-helper issue / Invalid credentials error.
// See https://github.com/wedeploy/cli/issues/323
// This passes the token as part of the remote address
// Security risk: prone to sniffing (on the same machine) on most operating systems.

package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/wedeploy/cli/deployment/internal/groupuid"

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
// 2.5.0 (Aug 18, 2015): weird git credential- error message, but still works
// 2.5.1 (Aug 28, 2015): weird git credential- error message, but still works
// 2.5.3 (Sep 18, 2015): starts breaking
// 2.13.3 (Jul 13, 2017): working again
const gitAffectedVersions = "> 2.5.1, < 2.13.3"

func (t *Transport) pushHack() (groupUID string, err error) {
	var params = []string{"push", t.getGitRemote(), "master", "--force"}

	if verbose.Enabled {
		params = append(params, "--verbose")
	}

	verbose.Debug(fmt.Sprintf("Running git push %v master -force",
		verbose.SafeEscape(t.getGitRemote())))

	var wectx = t.settings.ConfigContext

	var cmd = exec.CommandContext(t.ctx, "git", params...)
	cmd.Env = append(t.getConfigEnvs(),
		"GIT_TERMINAL_PROMPT=0",
		envs.GitCredentialRemoteToken+"="+wectx.Token(),
	)
	cmd.Dir = t.settings.WorkDir

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
		case strings.Contains(bs, "fatal: Authentication failed for"),
			strings.Contains(bs, "could not read Username"):
			return "", errors.New("invalid credentials: please update git and try again http://git-scm.com")
		case strings.Contains(bs, "error: "):
			return "", fmt.Errorf("git error: %v", verbose.SafeEscape(getGitErrors(bs).Error()))
		default:
			return "", err
		}
	}

	return groupuid.Extract(bufErr.String())
}

func (t *Transport) addRemoteHack() error {
	verbose.Debug("Adding remote with token")
	var wectx = t.settings.ConfigContext

	var gitServer = fmt.Sprintf("https://%v:@git.%v/%v.git",
		wectx.Token(),
		wectx.InfrastructureDomain(),
		t.settings.ProjectID)

	var params = []string{"remote", "add", t.getGitRemote(), gitServer}

	verbose.Debug(fmt.Sprintf("Running git remote add %v %v",
		t.getGitRemote(),
		verbose.SafeEscape(gitServer)))

	var cmd = exec.CommandContext(t.ctx, "git", params...)
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir

	if verbose.IsUnsafeMode() {
		cmd.Stderr = errStream
	}

	return cmd.Run()
}

func (t *Transport) useCredentialHack() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	v, err := version.NewVersion(t.gitVersion)

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
		verbose.Debug("git version " + t.gitVersion + " is not compatible with credential-helper due to a bug")
		verbose.Debug("fall back to passing token on remote")
		verbose.Debug("limited debug messages for security reasons")
		verbose.Debug("updating git is highly recommended")
	}

	return p
}
