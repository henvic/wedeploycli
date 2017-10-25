package deployment

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/envs"
	"github.com/wedeploy/cli/userhome"
	"github.com/wedeploy/cli/verbose"
)

// Cleanup directory
func (d *Deploy) Cleanup() error {
	return os.RemoveAll(d.getGitPath())
}

// CreateGitDirectory creates the git directory for the deployment
func (d *Deploy) CreateGitDirectory() error {
	return os.MkdirAll(d.getGitPath(), 0700)
}

func (d *Deploy) getConfigEnvs() (es []string) {
	if len(d.gitEnvCache) != 0 {
		return d.gitEnvCache
	}

	var home = filepath.Join(userhome.GetHomeDir(), ".wedeploy")
	var originals = os.Environ()
	var envs = map[string]string{}

	for _, o := range originals {
		if e := strings.SplitN(o, "=", 2); len(e) == 2 {
			envs[e[0]] = e[1]
		}
	}

	envs["GIT_DIR"] = d.getGitPath()
	envs["GIT_WORK_TREE"] = d.Path

	switch runtime.GOOS {
	case "windows":
		verbose.Debug("Microsoft Windows detected: using git system config")
	default:
		envs["GIT_CONFIG_NOSYSTEM"] = "true"
	}

	envs["HOME"] = home
	envs["XDG_CONFIG_HOME"] = home
	envs["GIT_CONFIG"] = filepath.Join(d.getGitPath(), "config")

	for key, value := range envs {
		if !strings.HasPrefix(key, fmt.Sprintf("%s=", key)) {
			es = append(es, fmt.Sprintf("%s=%s", key, value))
		}
	}

	d.gitEnvCache = es
	return es
}

// InitializeRepository as a git repo
func (d *Deploy) InitializeRepository() error {
	// preload the config envs before proceeding (just for the verbose msg)
	_ = d.getConfigEnvs()

	if err := d.getGitVersion(); err != nil {
		return err
	}

	var params = []string{"init"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path
	cmd.Stderr = errStream

	if err := cmd.Run(); err != nil {
		return err
	}

	if err := d.setKeepLineEndings(); err != nil {
		return err
	}

	return d.setGitAuthor()
}

func (d *Deploy) getGitVersion() error {
	var params = []string{"version"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path
	var buf bytes.Buffer
	cmd.Stderr = errStream
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return err
	}

	verbose.Debug(buf.String())

	// filter using semver partially
	r := regexp.MustCompile(`(\d+.\d+.\d+)(-[0-9A-Za-z-]*.\d*)?`)
	var b = r.FindStringSubmatch(buf.String())

	switch len(b) {
	case 0:
		d.gitVersion = buf.String()
	default:
		d.gitVersion = b[0]
	}

	return nil
}

func (d *Deploy) setKeepLineEndings() error {
	var params = []string{"config", "core.autocrlf", "false", "--local"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path
	cmd.Stderr = errStream

	return cmd.Run()
}

func (d *Deploy) getGitPath() string {
	var cachePath = d.Path

	// on Windows, drive units start with ":", if we don't remove it we get the following error:
	// 'GetFileAttributesEx U:\...U:\...: The filename, directory name, or volume label syntax is incorrect.
	if runtime.GOOS == "windows" {
		cachePath = strings.Replace(d.Path, ":", "", 1)
	}

	return filepath.Join(userhome.GetHomeDir(), ".wedeploy", "tmp", "repos", cachePath)
}

func (d *Deploy) getGitRemote() string {
	var remote = d.ConfigContext.Remote()

	// always add a "wedeploy-" prefix to all deployment remote endpoints, but "wedeploy"
	if remote != "wedeploy" {
		remote = "wedeploy" + "-" + remote
	}

	return remote
}

func (d *Deploy) setGitAuthor() error {
	if err := d.setGitAuthorName(); err != nil {
		return err
	}

	return d.setGitAuthorEmail()
}

func (d *Deploy) setGitAuthorName() error {
	var params = []string{"config", "user.name", "WeDeploy user", "--local"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path
	cmd.Stderr = errStream

	return cmd.Run()
}

func (d *Deploy) setGitAuthorEmail() error {
	var params = []string{"config", "user.email", "user@deployment", "--local"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path
	cmd.Stderr = errStream

	return cmd.Run()
}

// GetCurrentBranch gets the current branch
func (d *Deploy) GetCurrentBranch() (branch string, err error) {
	var params = []string{"symbolic-ref", "HEAD"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
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

func (d *Deploy) stageEachService(path string) error {
	var params = []string{"add", path}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path
	cmd.Stderr = errStream

	return cmd.Run()
}

func (d *Deploy) stageAllFiles() (err error) {
	for _, s := range d.Services {
		if err := d.stageEachService(s.Location); err != nil {
			return err
		}
	}

	if err = d.maybeRenameServiceIDs(); err != nil {
		return errwrap.Wrapf("can't stage custom wedeploy.json to replace service ID: {{err}}", err)
	}

	return nil
}

func (d *Deploy) maybeRenameServiceIDs() error {
	for _, remapLocation := range d.LocationRemap {
		for _, service := range d.Services {
			if service.Location != remapLocation {
				continue
			}

			if err := d.renameServiceID(service); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Deploy) gitRenameServiceID(content []byte, path string) error {
	switch hashObject, err := d.gitRenameServiceIDHashObject(content); {
	case err != nil:
		return err
	default:
		return d.gitRenameServiceIDUpdateIndex(hashObject, path)
	}
}

func (d *Deploy) gitRenameServiceIDHashObject(content []byte) (hashObject string, err error) {
	var params = []string{"hash-object", "-w", "--stdin"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path
	var in = &bytes.Buffer{}
	var out = &bytes.Buffer{}
	cmd.Stdin = in
	cmd.Stderr = errStream
	cmd.Stdout = out

	verbose.Debug(fmt.Sprintf("Using hash-object:\n%v", string(content)))

	if _, err := in.Write(content); err != nil {
		return "", err
	}

	if err = cmd.Run(); err != nil {
		return "", err
	}

	return out.String(), nil
}

func (d *Deploy) gitRenameServiceIDUpdateIndex(hashObject, path string) error {
	var params = []string{"update-index", "--add", "--cacheinfo", "100644", hashObject, path}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	return cmd.Run()
}

func (d *Deploy) getLastCommit() (commit string, err error) {
	var params = []string{"rev-parse", "HEAD"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	var buf bytes.Buffer
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = &buf

	err = cmd.Run()

	if err != nil {
		return "", errwrap.Wrapf("can't get last commit: {{err}}", err)
	}

	commit = strings.TrimSpace(buf.String())
	return commit, nil
}

// Commit adds all files and commits
func (d *Deploy) Commit() (commit string, err error) {
	if err = d.stageAllFiles(); err != nil {
		return "", err
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
	cmd.Env = d.getConfigEnvs()
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

// Push deployment to the WeDeploy remote
func (d *Deploy) Push() (groupUID string, err error) {
	if d.useGitCredentialHack() {
		return d.pushHack()
	}

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
			return "", errors.New("Invalid credentials")
		case strings.Contains(bs, "error: "):
			return "", getGitErrors(bs)
		default:
			return "", err
		}
	}

	return tryGetPushGroupUID(*bufErr)
}

// AddRemote on project
func (d *Deploy) AddRemote() error {
	if d.useGitCredentialHack() {
		return d.addRemoteHack()
	}

	var gitServer = fmt.Sprintf("https://git.%v/%v.git",
		d.ConfigContext.InfrastructureDomain(),
		d.ProjectID)

	var params = []string{"remote", "add", d.getGitRemote(), gitServer}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	return cmd.Run()
}

func (d *Deploy) addEmptyCredentialHelper() (err error) {
	// If credential.helper is configured to the empty string, this resets the helper list to empty
	// (so you may override a helper set by a lower-priority config file by configuring the empty-string helper,
	// followed by whatever set of helpers you would like).
	// https://www.kernel.org/pub/software/scm/git/docs/gitcredentials.html
	var params = []string{"config", "--add", "credential.helper", ""}
	verbose.Debug("Resetting credential helpers")
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = outStream
	return cmd.Run()
}

func (d *Deploy) addCredentialHelper() (err error) {
	if d.useGitCredentialHack() {
		verbose.Debug("Skipping adding git credential helper")
		return nil
	}

	if err := d.addEmptyCredentialHelper(); err != nil {
		return err
	}

	bin, err := getWeExecutable()

	if err != nil {
		return err
	}

	var credentialHelper = bin + " git-credential-helper"

	// Windows... Really? Really? Really? Really.
	if runtime.GOOS == "windows" {
		credentialHelper = strings.Replace(credentialHelper, `\`, `\\`, -1)
	}

	var params = []string{"config", "--add", "credential.helper", credentialHelper}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(d.Context, "git", params...)
	cmd.Env = d.getConfigEnvs()
	cmd.Dir = d.Path
	cmd.Stderr = errStream
	cmd.Stdout = outStream
	return cmd.Run()
}
