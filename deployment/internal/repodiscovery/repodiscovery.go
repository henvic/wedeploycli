package repodiscovery

import (
	"path/filepath"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/services"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Repository found.
type Repository struct {
	Services          []string `json:"repos,omitempty"`
	Path              string   `json:"path,omitempty"`
	Origin            string   `json:"origins,omitempty"`
	Commit            string   `json:"commit,omitempty"`
	CommitAuthor      string   `json:"commitAuthor,omitempty"`
	CommitAuthorEmail string   `json:"commitAuthorEmail,omitempty"`
	CommitMessage     string   `json:"commitMessage,omitempty"`
	CommitDate        string   `json:"commitDate,omitempty"`
	Branch            string   `json:"branch,omitempty"`
	CleanWorkingTree  bool     `json:"cleanWorkingTree,omitempty"`
}

// Discover repositories
type Discover struct {
	Path     string
	Services services.ServiceInfoList

	config *config.Config
	head   *plumbing.Reference
}

// Run discovery process
func (d *Discover) Run() (repos []Repository, repoless []string, err error) {
	d.Path, err = filepath.Abs(d.Path)

	if err != nil {
		return nil, nil, err
	}

	repoOnRoot, err := d.walkFn(d.Path, d.Services.GetIDs())

	if err != nil && err != git.ErrRepositoryNotExists {
		return nil, nil, err
	}

	if repoOnRoot != nil {
		repos = append(repos, *repoOnRoot)
		return repos, repoless, nil
	}

	for _, serviceInfo := range d.Services {
		repo, err := d.walkFn(serviceInfo.Location, []string{serviceInfo.ServiceID})

		if err == git.ErrRepositoryNotExists {
			repoless = append(repoless, serviceInfo.ServiceID)
			continue
		}

		if err != nil {
			return nil, nil, errwrap.Wrapf("error discovering git repos for "+serviceInfo.ServiceID+": {{err}}", err)
		}

		repos = append(repos, *repo)
	}

	return repos, repoless, nil
}

func (d *Discover) walkFn(path string, services []string) (*Repository, error) {
	repo, err := git.PlainOpen(path)

	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()

	if err != nil {
		return nil, err
	}

	status, err := worktree.Status()

	if err != nil {
		return nil, err
	}

	d.head, err = repo.Head()

	if err != nil {
		return nil, err
	}

	d.config, err = repo.Config()

	if err != nil {
		return nil, err
	}

	var branch, remote = d.maybeGetBranchAndRemote()

	commitHash := d.head.Hash()
	commit, err := repo.CommitObject(commitHash)

	if err != nil {
		return nil, err
	}

	rel, err := filepath.Rel(d.Path, path)

	if err != nil {
		return nil, err
	}

	if rel == "." {
		rel = ""
	}

	return &Repository{
		Services:          services,
		Path:              rel,
		Origin:            d.maybeGetOriginURL(remote),
		Commit:            commitHash.String(),
		CommitAuthor:      commit.Author.Name,
		CommitAuthorEmail: commit.Author.Email,
		CommitMessage:     commit.Message,
		CommitDate:        commit.Author.When.String(),
		Branch:            branch,
		CleanWorkingTree:  isWorkingTreeClean(status),
	}, nil
}

func (d *Discover) maybeGetBranchAndRemote() (branch, remote string) {
	if !d.head.Name().IsBranch() {
		return
	}

	branch = d.head.Name().Short()

	if b, ok := d.config.Branches[branch]; ok {
		remote = b.Remote
	}

	return
}

func (d *Discover) maybeGetOriginURL(remote string) string {
	r, ok := d.config.Remotes[remote]

	if !ok {
		return ""
	}

	var candidates = r.URLs

	for _, eu := range candidates {
		if gitRepoURL, err := ExtractRepoURL(eu); err == nil {
			return gitRepoURL
		}
	}

	return ""
}

func isWorkingTreeClean(s git.Status) bool {
	for _, status := range s {
		if status.Staging == git.Unmodified && status.Worktree == git.Unmodified {
			continue
		}

		return false
	}

	return true
}
