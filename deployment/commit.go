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
	"github.com/wedeploy/cli/verbose"
)

// Commit adds all files and commits
func (d *Deploy) Commit() (commit string, err error) {
	if err = d.stageAllFiles(); err != nil {
		return "", err
	}

	msg, err := d.commitMessage()

	if err != nil {
		return "", err
	}

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

type deployInfo struct {
	CLIVersion   string                     `json:"cliVersion,omitempty"`
	Time         string                     `json:"time,omitempty"`
	OnlyBuild    bool                       `json:"onlyBuild,omitempty"` // @todo
	Repositories []repodiscovery.Repository `json:"repos,omitempty"`
	Repoless     []string                   `json:"repoless,omitempty"`
}

func (d *Deploy) commitMessage() (message string, err error) {
	date := time.Now().Format(time.RubyDate)

	template := `Deployment at %v

	---
	%v`

	return fmt.Sprintf(template, date, d.deployInfo()), err
}

func (d *Deploy) deployInfo() string {
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

	di := deployInfo{
		CLIVersion:   version,
		Time:         time.Now().Format(time.RubyDate),
		Repositories: repositories,
		Repoless:     repoless,
	}

	bdi, err := json.Marshal(di)

	if err != nil {
		verbose.Debug(err)
	}

	return string(bdi)
}
