package tiny

import (
	"github.com/wedeploy/cli/deployment/internal/repodiscovery"
)

// Info about the deployment.
// Copy of deployment.Info struct.
type Info struct {
	CLIVersion string `json:"cliVersion,omitempty"`
	Time       string `json:"time,omitempty"`
	Deploy     bool   `json:"deploy,omitempty"`

	Repositories []repodiscovery.Repository `json:"repos,omitempty"`
	Repoless     []string                   `json:"repoless,omitempty"`
}

// Tiny deployment info.
type Tiny struct {
	Deploy bool `json:"deploy"`

	Commit *commitInfo `json:"commit,omitempty"`
}

type commitInfo struct {
	SHA         string `json:"sha,omitempty"`
	Message     string `json:"message,omitempty"`
	AuthorName  string `json:"authorName,omitempty"`
	AuthorEmail string `json:"authorEmail,omitempty"`
	Date        string `json:"date,omitempty"`
}

// Convert deployinfo format to this tiny format.
// see https://github.com/wedeploy/nodegit/issues/43#issuecomment-417728174
func Convert(i Info) Tiny {
	var t = Tiny{
		Deploy: i.Deploy,
	}

	convertCommit(i, &t)

	return t
}

func convertCommit(i Info, t *Tiny) {
	if len(i.Repoless) != 0 || len(i.Repositories) != 1 {
		return
	}

	repo := i.Repositories[0]

	t.Commit = &commitInfo{
		SHA:         repo.Commit,
		Message:     repo.CommitMessage,
		AuthorName:  repo.CommitAuthor,
		AuthorEmail: repo.CommitAuthorEmail,
		Date:        repo.CommitDate,
	}
}