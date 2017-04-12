package deployment

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"os"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/userhome"
	"github.com/wedeploy/cli/verbose"
)

var (
	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr
)

// Deploy project
type Deploy struct {
	Context           context.Context
	ProjectID         string
	Path              string
	Remote            string
	RepoAuthorization string
	GitRemoteAddress  string
}

func (d *Deploy) getGitPath() string {
	return filepath.Join(userhome.GetHomeDir(), ".wedeploy", "tmp", "repos", d.Path)
}

func (d *Deploy) getGitRemote() string {
	var remote = d.Remote

	// always add a "wedeploy-" prefix to all deployment remote endpoints, but "wedeploy"
	if d.Remote != "wedeploy" {
		remote = "wedeploy" + "-" + d.Remote
	}

	return remote
}

// Cleanup directory
func (d *Deploy) Cleanup() error {
	return os.RemoveAll(d.getGitPath())
}

// CreateGitDirectory creates the git directory for the deployment
func (d *Deploy) CreateGitDirectory() error {
	return os.MkdirAll(d.getGitPath(), 0775)
}

// InitializeRepository as a git repo
func (d *Deploy) InitializeRepository() error {
	var params = []string{"init"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	cmd.Dir = d.Path
	cmd.Stderr = errStream

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// GetCurrentBranch gets the current branch
func (d *Deploy) GetCurrentBranch() (branch string, err error) {
	var params = []string{"symbolic-ref", "HEAD"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	var buf bytes.Buffer
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = &buf

	err = cmd.Run()

	if err != nil {
		return "", errwrap.Wrapf("Can not get current branch: {{err}}", err)
	}

	branch = strings.TrimPrefix(strings.TrimSpace(buf.String()), "refs/heads/")
	return branch, nil
}

func (d *Deploy) stageAllFiles() (err error) {
	var params = []string{"add", "."}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = outStream

	return cmd.Run()
}

func (d *Deploy) getLastCommit() (commit string, err error) {
	var params = []string{"rev-parse", "HEAD"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	var buf bytes.Buffer
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = &buf

	err = cmd.Run()

	if err != nil {
		return "", errwrap.Wrapf("Can not get last commit: {{err}}", err)
	}

	commit = strings.TrimSpace(buf.String())
	return commit, nil
}

// Commit adds all files and commits
func (d *Deploy) Commit() (commit string, err error) {
	if err = d.stageAllFiles(); err != nil {
		return "", errwrap.Wrapf("Trying to stage all files: {{err}}", err)
	}

	var msg = fmt.Sprintf("Deployment at %v", time.Now().Format(time.RubyDate))
	var params = []string{"commit", "--allow-empty", "--message", msg}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	cmd.Dir = d.Path

	if verbose.Enabled {
		cmd.Stderr = errStream
	}

	err = cmd.Run()

	if err != nil {
		return "", errwrap.Wrapf("Can not commit: {{err}}", err)
	}

	commit, err = d.getLastCommit()

	if err != nil {
		return "", err
	}

	verbose.Debug("commit", commit)
	return commit, nil
}

func (d *Deploy) verboseOnPush() {
	if !verbose.Enabled {
		return
	}

	verbose.Debug(color.Format(color.FgBlue, "Push Authorization") +
		color.Format(color.FgRed, ": ") +
		verbose.SafeEscape(d.RepoAuthorization))
}

func copyErrStreamAndVerbose(cmd *exec.Cmd) (bufErr bytes.Buffer) {
	if verbose.Enabled {
		cmd.Stderr = io.MultiWriter(&bufErr, os.Stderr)
	} else {
		cmd.Stderr = &bufErr
	}

	return bufErr
}

// Push deployment to the WeDeploy remote
func (d *Deploy) Push() error {
	var params = []string{"push", d.getGitRemote(), "master", "--force"}

	if verbose.Enabled {
		params = append(params, "--verbose")
	}

	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	if d.RepoAuthorization != "" {
		d.verboseOnPush()
		params = append([]string{
			"-c",
			"http." + d.GitRemoteAddress + ".extraHeader=Authorization: " + d.RepoAuthorization,
		}, params...)
	}

	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env,
		"GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path,
		"GIT_TERMINAL_PROMPT=0",
	)
	cmd.Dir = d.Path

	var bufErr = copyErrStreamAndVerbose(cmd)
	var err = cmd.Run()

	if err != nil && strings.Contains(bufErr.String(), "could not read Username") {
		return errors.New("Invalid credentials")
	}

	return err
}

// AddRemote on project
func (d *Deploy) AddRemote() error {
	var params = []string{"remote", "add", d.getGitRemote(), d.GitRemoteAddress}
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = outStream
	return cmd.Run()
}

func (d *Deploy) cleanupAfter() {
	if err := d.Cleanup(); err != nil {
		verbose.Debug(
			errwrap.Wrapf("Error trying to clean up directory after deployment: {{err}}", err))
	}
}

// Do deployment
func (d *Deploy) Do() error {
	if err := d.Cleanup(); err != nil {
		return errwrap.Wrapf("Can not clean up directory for deployment: {{err}}", err)
	}

	if err := d.CreateGitDirectory(); err != nil {
		return errwrap.Wrapf("Can not create temporary directory for deployment: {{err}}", err)
	}

	defer d.cleanupAfter()

	if err := d.InitializeRepository(); err != nil {
		return err
	}

	if _, err := d.Commit(); err != nil {
		return err
	}

	if err := d.AddRemote(); err != nil {
		return err
	}

	if err := d.Push(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return errwrap.Wrapf("Can not deploy (push failure)", err)
		}

		return errwrap.Wrapf("Unexpected push failure: can not deploy ({{err}})", err)
	}

	return nil
}
