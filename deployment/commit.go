package deployment

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/deployment/internal/repodiscovery"
	"github.com/wedeploy/cli/deployment/internal/repodiscovery/tiny"
	"github.com/wedeploy/cli/verbose"
)

// Commit adds all files and commits
func (d *Deploy) Commit() (commit string, err error) {
	if err = d.stageAllFiles(); err != nil {
		return "", err
	}

	msg := d.Info()

	var params = []string{
		"commit",
		"--allow-empty",
		"--message",
		msg,
	}

	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.ctx, "git", params...)
	cmd.Env = append(d.getConfigEnvs(), "GIT_WORK_TREE="+d.Path)
	cmd.Dir = d.Path

	if verbose.Enabled {
		cmd.Stderr = errStream
	}

	err = cmd.Run()

	if err != nil {
		return "", errwrap.Wrapf("can't commit: {{err}}", err)
	}

	commit, err = d.getLastCommit()

	if err != nil {
		return "", err
	}

	verbose.Debug("commit", commit)
	return commit, nil
}

// Info about the deployment.
type Info struct {
	CLIVersion string `json:"cliVersion,omitempty"`
	Time       string `json:"time,omitempty"`
	Deploy     bool   `json:"deploy,omitempty"`

	Repositories []repodiscovery.Repository `json:"repos,omitempty"`
	Repoless     []string                   `json:"repoless,omitempty"`
}

// Info about the deployment.
func (d *Deploy) Info() string {
	version := fmt.Sprintf("%s %s/%s",
		defaults.Version,
		runtime.GOOS,
		runtime.GOARCH)

	rd := repodiscovery.Discover{
		Path:     d.Path,
		Services: d.Services,
	}

	repositories, repoless, err := rd.Run()

	if err != nil {
		verbose.Debug(err)
		return ""
	}

	di := Info{
		CLIVersion:   version,
		Time:         time.Now().Format(time.RubyDate),
		Deploy:       !d.OnlyBuild,
		Repositories: repositories,
		Repoless:     repoless,
	}

	bdi, err := json.Marshal(tiny.Convert(tiny.Info(di)))

	if err != nil {
		verbose.Debug(err)
	}

	return string(bdi)
}
