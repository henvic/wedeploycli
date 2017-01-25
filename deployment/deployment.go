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
	"github.com/wedeploy/cli/verbose"
)

const (
	unixTimeFormat = "Mon Jan _2 15:04:05 MST 2006"
)

var (
	outStream io.Writer = os.Stdout
	errStream io.Writer = os.Stderr
)

// Deploy project
type Deploy struct {
	Context            context.Context
	ProjectID          string
	Path               string
	Force              bool
	Remote             string
	GitRemoteAddress   string
	uncommittedChanges bool
}

func (d *Deploy) getGitRemote() string {
	var remote = d.Remote

	// always add a "wedeploy-" prefix to all deployment remote endpoints, but "wedeploy"
	if d.Remote != "wedeploy" {
		remote = "wedeploy" + "-" + d.Remote
	}

	return remote
}

// InitalizeRepository as a git repo
func (d *Deploy) InitalizeRepository() error {
	var params = []string{"init"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Dir = d.Path
	cmd.Stderr = errStream

	fmt.Fprintf(outStream, "Initializing git repository on project folder\n")

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// InitalizeRepositoryIfNotExists as a git repo
func (d *Deploy) InitalizeRepositoryIfNotExists() error {
	switch _, err := os.Stat(filepath.Join(d.Path, ".git")); {
	case os.IsNotExist(err):
		return d.InitalizeRepository()
	case err != nil:
		return errwrap.Wrapf("Unexpected error when trying to find .git: {{err}}", err)
	default:
		return nil
	}
}

// GetCurrentBranch gets the current branch
func (d *Deploy) GetCurrentBranch() (branch string, err error) {
	var params = []string{"symbolic-ref", "HEAD"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
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

// CheckCurrentBranchIsMaster checks if the current branch is master
func (d *Deploy) CheckCurrentBranchIsMaster() error {
	var branch, err = d.GetCurrentBranch()

	if err != nil {
		return err
	}

	if branch != "master" {
		return errors.New("Current branch is not master")
	}

	return nil
}

// CheckUncommittedChanges checks for uncommitted changes on the staging area
func (d *Deploy) CheckUncommittedChanges() (stage string, err error) {
	var params = []string{"status", "--porcelain"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	var buf bytes.Buffer
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return "", err
	}

	if buf.Len() != 0 {
		fmt.Fprintf(outStream, "You have uncommitted changes\n")
		d.uncommittedChanges = true
	}

	return buf.String(), nil
}

func (d *Deploy) stageAllFiles() (err error) {
	var params = []string{"add", "."}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = outStream

	return cmd.Run()
}

func (d *Deploy) getLastCommit() (commit string, err error) {
	var params = []string{"rev-parse", "HEAD"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
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

	var msg = fmt.Sprintf("Deployment at %v", time.Now().Format(unixTimeFormat))
	var params = []string{"commit", "--message", msg}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Dir = d.Path
	cmd.Stderr = errStream

	err = cmd.Run()

	if err != nil {
		return "", errwrap.Wrapf("Can not commit: {{err}}", err)
	}

	commit, err = d.getLastCommit()

	if err != nil {
		return "", err
	}

	var shortCommit = commit

	if len(commit) > 7 {
		shortCommit = commit[0:7]
	}

	fmt.Fprintf(outStream, "commit %v: %v\n", shortCommit, msg)
	return commit, nil
}

// Push project to the WeDeploy remote
func (d *Deploy) Push() error {
	var params = []string{"push", d.getGitRemote(), "master"}

	if d.Force {
		params = append(params, "--force")
	}

	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = outStream

	return cmd.Run()
}

func (d *Deploy) removeRemote() error {
	var params = []string{"remote", "rm", d.getGitRemote()}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Dir = d.Path
	return cmd.Run()
}

// AddRemote on project
func (d *Deploy) AddRemote() error {
	_ = d.removeRemote()
	var params = []string{"remote", "add", d.getGitRemote(), d.GitRemoteAddress}
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = outStream
	return cmd.Run()
}
