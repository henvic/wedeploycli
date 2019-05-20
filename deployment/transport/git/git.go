package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/deployment/internal/groupuid"
	"github.com/wedeploy/cli/deployment/transport"
	"github.com/wedeploy/cli/envs"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/userhome"
	"github.com/wedeploy/cli/verbose"
)

var errStream io.Writer = os.Stderr

// Transport using go-git.
type Transport struct {
	ctx      context.Context
	settings transport.Settings

	start time.Time
	end   time.Time

	gitEnvCache []string
	gitVersion  string
}

// Stage files.
func (t *Transport) Stage(s services.ServiceInfoList) (err error) {
	verbose.Debug("Staging files")

	for _, service := range s {
		if err = t.stageService(filepath.Base(service.Location)); err != nil {
			return err
		}
	}

	return nil
}

func (t *Transport) stageService(dest string) error {
	var params = []string{"add", dest}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir
	cmd.Stderr = errStream

	return cmd.Run()
}

// Commit adds all files and commits
func (t *Transport) Commit(message string) (commit string, err error) {
	var params = []string{
		"commit",
		"--no-verify",
		"--allow-empty",
		"--message",
		message,
	}

	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir

	if verbose.Enabled {
		cmd.Stderr = errStream
	}

	err = cmd.Run()

	if err != nil {
		return "", errwrap.Wrapf("can't commit: {{err}}", err)
	}

	commit, err = t.getLastCommit()

	if err != nil {
		return "", err
	}

	verbose.Debug("commit", commit)
	return commit, nil
}

func (t *Transport) getLastCommit() (commit string, err error) {
	var params = []string{"rev-parse", "HEAD"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	var buf bytes.Buffer
	cmd.Dir = t.settings.WorkDir
	cmd.Stderr = errStream
	cmd.Stdout = &buf

	err = cmd.Run()

	if err != nil {
		return "", errwrap.Wrapf("can't get last commit: {{err}}", err)
	}

	commit = strings.TrimSpace(buf.String())
	return commit, nil
}

// Push deployment to the Liferay Cloud remote
func (t *Transport) Push() (groupUID string, err error) {
	t.start = time.Now()
	defer func() {
		t.end = time.Now()
	}()

	if t.useCredentialHack() {
		return t.pushHack()
	}

	var params = []string{"push", t.getGitRemote(), "master", "--force", "--no-verify"}

	if verbose.Enabled {
		params = append(params, "--verbose")
	}

	var wectx = t.settings.ConfigContext

	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = append(t.getConfigEnvs(),
		"GIT_TERMINAL_PROMPT=0",
		envs.GitCredentialRemoteToken+"="+wectx.Token(),
	)
	cmd.Dir = t.settings.WorkDir

	var bufErr = copyErrStreamAndVerbose(cmd)
	err = cmd.Run()

	if err != nil {
		bs := bufErr.String()
		switch {
		case strings.Contains(bs, "fatal: Authentication failed for"),
			strings.Contains(bs, "could not read Username"):
			return "", errors.New("invalid credentials when pushing deployment")
		case strings.Contains(bs, "error: "):
			return "", getGitErrors(bs)
		default:
			return "", err
		}
	}

	return groupuid.Extract(bufErr.String())
}

// UploadDuration for deployment (only correct after it finishes)
func (t *Transport) UploadDuration() time.Duration {
	return t.end.Sub(t.start)
}

// Setup as a git repo
func (t *Transport) Setup(ctx context.Context, settings transport.Settings) error {
	t.ctx = ctx
	t.settings = settings

	if hasGit := existsDependency("git"); !hasGit {
		return errors.New("git was not found on your system: please visit https://git-scm.com/")
	}

	// preload the config envs
	_ = t.getConfigEnvs()

	if err := t.getGitVersion(); err != nil {
		return err
	}

	return nil
}

func existsDependency(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (t *Transport) getGitVersion() error {
	var params = []string{"version"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir
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
		t.gitVersion = buf.String()
	default:
		t.gitVersion = b[0]
	}

	return nil
}

// Init repository
func (t *Transport) Init() (err error) {
	var params = []string{"init"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir
	cmd.Stderr = errStream

	if err := cmd.Run(); err != nil {
		return err
	}

	if err := t.setKeepLineEndings(); err != nil {
		return err
	}

	if err := t.setStopLineEndingsWarnings(); err != nil {
		return err
	}

	return t.setGitAuthor()
}

func (t *Transport) setKeepLineEndings() error {
	var params = []string{"config", "core.autocrlf", "false", "--local"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir
	cmd.Stderr = errStream

	return cmd.Run()
}

func (t *Transport) setStopLineEndingsWarnings() error {
	var params = []string{"config", "core.safecrlf", "false", "--local"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir
	cmd.Stderr = errStream

	return cmd.Run()
}

func (t *Transport) setGitAuthor() error {
	if err := t.setGitAuthorName(); err != nil {
		return err
	}

	return t.setGitAuthorEmail()
}

func (t *Transport) setGitAuthorName() error {
	var params = []string{"config", "user.name", "Liferay Cloud user", "--local"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir
	cmd.Stderr = errStream

	return cmd.Run()
}

func (t *Transport) setGitAuthorEmail() error {
	var params = []string{"config", "user.email", "user@deployment", "--local"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir
	cmd.Stderr = errStream

	return cmd.Run()
}

func (t *Transport) getGitRemote() string {
	var remote = t.settings.ConfigContext.Remote()

	// always add a "wedeploy-" prefix to all deployment remote endpoints, but "lcp"
	if remote != "lcp" {
		remote = "lcp" + "-" + remote
	}

	return remote
}

// ProcessIgnored gets what file should be ignored.
func (t *Transport) ProcessIgnored() (map[string]struct{}, error) {
	var params = []string{"status", "--ignored", "--untracked-files=all", "--porcelain", "--", "."}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = append(t.getConfigEnvs(), "GIT_WORK_TREE="+t.settings.Path)
	cmd.Dir = t.settings.Path
	cmd.Stderr = errStream

	var out = &bytes.Buffer{}
	cmd.Stdout = out
	var list = map[string]struct{}{}

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	const ignorePattern = "!! "

	for _, w := range bytes.Split(out.Bytes(), []byte("\n")) {
		if bytes.HasPrefix(w, []byte(ignorePattern)) {
			p := filepath.Join(t.settings.Path,
				string(bytes.TrimPrefix(w, []byte(ignorePattern))))
			list[p] = struct{}{}
		}
	}

	if len(list) != 0 {
		verbose.Debug(fmt.Sprintf(
			"Ignoring %d files and directories found on .gitignore files",
			len(list)))

	}

	return list, nil
}

// AddRemote on project
func (t *Transport) AddRemote() (err error) {
	if t.useCredentialHack() {
		return t.addRemoteHack()
	}

	wectx := t.settings.ConfigContext

	var gitServer = fmt.Sprintf("https://git.%v/%v.git",
		wectx.InfrastructureDomain(),
		t.settings.ProjectID)

	var params = []string{"remote", "add", t.getGitRemote(), gitServer}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir
	cmd.Stderr = errStream

	if err = cmd.Run(); err != nil {
		return err
	}

	return t.addCredentialHelper()
}

func (t *Transport) addEmptyCredentialHelper() (err error) {
	// If credential.helper is configured to the empty string, this resets the helper list to empty
	// (so you may override a helper set by a lower-priority config file by configuring the empty-string helper,
	// followed by whatever set of helpers you would like).
	// https://www.kernel.org/pub/software/scm/git/docs/gitcredentials.html
	var params = []string{"config", "--add", "credential.helper", ""}
	verbose.Debug("Resetting credential helpers")
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir
	cmd.Stderr = errStream
	return cmd.Run()
}

func (t *Transport) addCredentialHelper() error {
	if t.useCredentialHack() {
		verbose.Debug("Skipping adding git credential helper")
		return nil
	}

	if err := t.addEmptyCredentialHelper(); err != nil {
		return err
	}

	bin, err := getWeExecutable()

	if err != nil {
		return err
	}

	// Windows... Really? Really? Really? Really.
	// See issue #323
	if runtime.GOOS == "windows" {
		bin = strings.Replace(bin, `\`, `/`, -1)
		bin = strings.Replace(bin, ` `, `\ `, -1)
	}

	var credentialHelper = bin + " git-credential-helper"

	var params = []string{"config", "--add", "credential.helper", credentialHelper}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir
	cmd.Stderr = errStream
	return cmd.Run()
}

func getWeExecutable() (string, error) {
	var exec, err = os.Executable()

	if err != nil {
		verbose.Debug(fmt.Sprintf("%v; falling back to os.Args[0]", err))
		return filepath.Abs(os.Args[0])
	}

	return exec, nil
}

// // filter using semver partially
var semverMatcher = regexp.MustCompile(`(\d+.\d+.\d+)(-[0-9A-Za-z-]*.\d*)?`)

// UserAgent of the transport layer.
func (t *Transport) UserAgent() string {
	var params = []string{"version"}
	verbose.Debug(fmt.Sprintf("Running git %v", strings.Join(params, " ")))
	var cmd = exec.CommandContext(t.ctx, "git", params...) // #nosec
	cmd.Env = t.getConfigEnvs()
	cmd.Dir = t.settings.WorkDir
	var buf bytes.Buffer
	cmd.Stderr = errStream
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		verbose.Debug(err)
		return "unknown"
	}

	var v = buf.String()
	verbose.Debug(v)

	var b = semverMatcher.FindStringSubmatch(v)

	if len(b) != 0 {
		return b[0]
	}

	return v
}

func (t *Transport) getConfigEnvs() (es []string) {
	if len(t.gitEnvCache) != 0 {
		return t.gitEnvCache
	}

	var originals = os.Environ()
	var vars = map[string]string{}

	for _, o := range originals {
		if e := strings.SplitN(o, "=", 2); len(e) == 2 {
			vars[e[0]] = e[1]
		}
	}

	if v, ok := vars[envs.SkipTLSVerification]; ok {
		vars["GIT_SSL_NO_VERIFY"] = v
	}

	var gitDir = filepath.Join(t.settings.WorkDir, ".git")

	vars["GIT_DIR"] = gitDir

	switch runtime.GOOS {
	case "windows":
		verbose.Debug("Microsoft Windows detected: using git system config")
	default:
		vars["GIT_CONFIG_NOSYSTEM"] = "true"
	}

	var sandboxHome = filepath.Join(userhome.GetHomeDir(), ".wedeploy", "git-sandbox")
	vars["HOME"] = sandboxHome
	vars["XDG_CONFIG_HOME"] = sandboxHome
	vars["GIT_CONFIG"] = filepath.Join(gitDir, "config")
	vars["GIT_WORK_TREE"] = t.settings.WorkDir

	for key, value := range vars {
		if !strings.HasPrefix(key, fmt.Sprintf("%s=", key)) {
			es = append(es, fmt.Sprintf("%s=%s", key, value))
		}
	}

	t.gitEnvCache = es
	return es
}

func copyErrStreamAndVerbose(cmd *exec.Cmd) *bytes.Buffer {
	var bufErr bytes.Buffer
	cmd.Stderr = &bufErr

	switch {
	case verbose.Enabled && verbose.IsUnsafeMode():
		cmd.Stderr = io.MultiWriter(&bufErr, os.Stderr)
	case verbose.Enabled:
		verbose.Debug(fmt.Sprintf(
			"Use %v=true to override security protection (see wedeploy/cli #327)",
			envs.UnsafeVerbose))
	}

	return &bufErr
}

func getGitErrors(s string) error {
	var parts = strings.Split(s, "\n")
	var list = []string{}
	for _, p := range parts {
		if strings.Contains(p, "error: ") {
			list = append(list, p)
		}
	}

	if len(list) == 0 {
		return nil
	}

	return fmt.Errorf("push: %v", strings.Join(list, "\n"))
}
