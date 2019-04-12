package gogit

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/deployment/internal/groupuid"
	"github.com/wedeploy/cli/deployment/transport"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/verbose"
	git "gopkg.in/src-d/go-git.v4"
	gitconfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	gotransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// Transport using go-git.
type Transport struct {
	ctx      context.Context
	settings transport.Settings

	repository *git.Repository
	worktree   *git.Worktree

	start time.Time
	end   time.Time
}

// Stage files.
func (t *Transport) Stage(s services.ServiceInfoList) (err error) {
	verbose.Debug("Staging files")

	for _, s := range s {
		if _, err = t.worktree.Add(filepath.Base(s.Location)); err != nil {
			return err
		}
	}

	return nil
}

// Commit adds all files and commits
func (t *Transport) Commit(message string) (commit string, err error) {
	verbose.Debug("Committing changes.")

	var hash plumbing.Hash

	hash, err = t.worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Liferay user",
			Email: "user@deployment",
			When:  time.Now(),
		},
	})

	if err != nil {
		return "", errwrap.Wrapf("can't commit: {{err}}", err)
	}

	commit = hash.String()

	verbose.Debug("commit", commit)
	return commit, nil
}

// Push deployment to the remote
func (t *Transport) Push() (groupUID string, err error) {
	verbose.Debug("Started uploading deployment package to the infrastructure")

	t.start = time.Now()
	defer func() {
		t.end = time.Now()
	}()

	buf := &bytes.Buffer{}

	o := &git.PushOptions{
		RemoteName: t.getGitRemote(),
		RefSpecs: []gitconfig.RefSpec{
			"+refs/heads/master:refs/heads/master",
		},
		Auth: &http.BasicAuth{
			Username: t.settings.ConfigContext.Token(),
		},
		Progress: buf,
	}

	err = t.repository.PushContext(t.ctx, o)

	switch {
	case err == gotransport.ErrAuthenticationRequired:
		return "", errwrap.Wrapf("Invalid credentials", err)
	case err != nil:
		return "", err
	}

	return groupuid.Extract(buf.String())
}

// UploadDuration for deployment (only correct after it finishes)
func (t *Transport) UploadDuration() time.Duration {
	return t.end.Sub(t.start)
}

// Setup as a git repo
func (t *Transport) Setup(ctx context.Context, settings transport.Settings) error {
	t.ctx = ctx
	t.settings = settings
	return nil
}

// Init repository
func (t *Transport) Init() (err error) {
	if t.repository, err = git.PlainInit(t.settings.WorkDir, false); err != nil {
		return errwrap.Wrapf("cannot create repository: {{err}}", err)
	}

	if t.worktree, err = t.repository.Worktree(); err != nil {
		return errwrap.Wrapf("cannot create worktree: {{err}}", err)
	}

	return nil
}

func (t *Transport) getGitRemote() string {
	var remote = t.settings.ConfigContext.Remote()

	// always add a "wedeploy-" prefix to all deployment remote endpoints, but "liferay"
	if remote != "liferay" {
		remote = "liferay" + "-" + remote
	}

	return remote
}

// ProcessIgnored gets what file should be ignored.
func (t *Transport) ProcessIgnored() (map[string]struct{}, error) {
	i := ignoreChecker{
		path: t.settings.Path,
	}

	return i.Process()
}

// AddRemote on project
func (t *Transport) AddRemote() (err error) {
	var gitServer = fmt.Sprintf("https://git.%v/%v.git",
		t.settings.ConfigContext.InfrastructureDomain(),
		t.settings.ProjectID)

	_, err = t.repository.CreateRemote(&gitconfig.RemoteConfig{
		Name: t.getGitRemote(),
		URLs: []string{gitServer},
	})

	return err
}

// UserAgent of the transport layer.
func (t *Transport) UserAgent() string {
	return "go-git v4"
}
