package deployment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
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
	AuthorEmail       string
	ProjectID         string
	Path              string
	Remote            string
	RepoAuthorization string
	GitRemoteAddress  string
	groupUID          string
	pushStartTime     time.Time
	pushEndTime       time.Time
	notify            Notifier
}

// Notifier interface
type Notifier func(string)

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

func (d *Deploy) unstageProjectJSON() (err error) {
	var params = []string{"reset", "--", "project.json"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	cmd.Dir = d.Path
	cmd.Stderr = errStream

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

	if err = d.unstageProjectJSON(); err != nil {
		return "", errwrap.Wrapf("Trying to unstage project.json: {{err}}", err)
	}

	var msg = fmt.Sprintf("Deployment at %v", time.Now().Format(time.RubyDate))

	var params = []string{
		"commit",
		"--allow-empty",
		"--message",
		msg,
	}

	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env, "GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path)
	cmd.Env = append(cmd.Env, fmt.Sprintf("GIT_AUTHOR_EMAIL=%v", d.AuthorEmail))
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

func copyErrStreamAndVerbose(cmd *exec.Cmd) *bytes.Buffer {
	var bufErr bytes.Buffer
	if verbose.Enabled {
		cmd.Stderr = io.MultiWriter(&bufErr, os.Stderr)
	} else {
		cmd.Stderr = &bufErr
	}

	return &bufErr
}

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
	if d.RepoAuthorization != "" {
		d.verboseOnPush()
		params = append([]string{
			"-c",
			"https." + d.GitRemoteAddress + ".extraHeader=Authorization: " + d.RepoAuthorization,
		}, params...)
	}

	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = append(cmd.Env,
		"GIT_DIR="+d.getGitPath(), "GIT_WORK_TREE="+d.Path,
		"GIT_TERMINAL_PROMPT=0",
	)
	cmd.Dir = d.Path

	var bufErr = copyErrStreamAndVerbose(cmd)
	err = cmd.Run()

	if err != nil && strings.Contains(bufErr.String(), "could not read Username") {
		return "", errors.New("Invalid credentials")
	}

	if err != nil {
		return "", err
	}

	return tryGetPushGroupUID(*bufErr)
}

var (
	gitRemoteDeployPrefix      = []byte("remote: deploy=")
	gitRemoteDeployErrorPrefix = []byte("remote: deployError=")
)

func tryGetPushGroupUID(buff bytes.Buffer) (groupUID string, err error) {
	for {
		line, err := buff.ReadBytes('\n')

		if bytes.HasPrefix(line, gitRemoteDeployPrefix) {
			return extractGroupUIDFromBuild(bytes.TrimPrefix(line, gitRemoteDeployPrefix))
		}

		if bytes.HasPrefix(line, gitRemoteDeployErrorPrefix) {
			return "", extractErrorFromBuild(bytes.TrimPrefix(line, gitRemoteDeployErrorPrefix))
		}

		if err == io.EOF {
			return "", errors.New("can't find deployment group UID response")
		}
	}
}

func extractErrorFromBuild(e []byte) error {
	var af apihelper.APIFault
	if errJSON := json.Unmarshal(e, &af); errJSON != nil {
		return fmt.Errorf(`can't process error message: "%s"`, e)
	}

	return af
}

type buildDeploymentOnGitServer struct {
	GroupUID string `json:"groupUid"`
}

func extractGroupUIDFromBuild(e []byte) (groupUID string, err error) {
	var bds []buildDeploymentOnGitServer

	if errJSON := json.Unmarshal(e, &bds); errJSON != nil {
		return "", errwrap.Wrapf("deployment response is invalid: {{err}}", errJSON)
	}

	if len(bds) == 0 {
		return "", errors.New("found no build during deployment")
	}

	return bds[0].GroupUID, nil
}

// AddRemote on project
func (d *Deploy) AddRemote() error {
	var params = []string{"remote", "add", d.getGitRemote(), d.GitRemoteAddress}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
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

// DiscardNotifier receiver
func DiscardNotifier(s string) {}

// Do deployment
func (d *Deploy) Do(n Notifier) (err error) {
	d.notify = n

	d.notify("Initializing deployment process")

	if err = d.Cleanup(); err != nil {
		return errwrap.Wrapf("Can not clean up directory for deployment: {{err}}", err)
	}

	if err = d.CreateGitDirectory(); err != nil {
		return errwrap.Wrapf("Can not create temporary directory for deployment: {{err}}", err)
	}

	defer d.cleanupAfter()

	if err = d.InitializeRepository(); err != nil {
		return err
	}

	d.notify("Preparing package…")

	if _, err = d.Commit(); err != nil {
		return err
	}

	if err = d.AddRemote(); err != nil {
		return err
	}

	d.notify("Uploading package…")

	if d.groupUID, err = d.Push(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return errwrap.Wrapf("deployment push failed", err)
		}

		return errwrap.Wrapf("deployment failed: {{err}}", err)
	}

	return nil
}

// GetGroupUID gets the deployment group UID
func (d *Deploy) GetGroupUID() string {
	return d.groupUID
}

// UploadDuration for deployment (only correct after it finishes)
func (d *Deploy) UploadDuration() time.Duration {
	return d.pushEndTime.Sub(d.pushStartTime)
}
