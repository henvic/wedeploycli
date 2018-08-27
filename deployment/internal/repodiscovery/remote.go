package repodiscovery

import (
	"errors"
	"net/url"
	"strings"
)

// ExtractRepoURL from remote
func ExtractRepoURL(remoteURL string) (repoURL string, err error) {
	if len(remoteURL) == 0 {
		return "", errors.New("empty remote")
	}

	if remoteURL[0] == '/' {
		return "", errors.New("origin is on the same machine")
	}

	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	if strings.HasPrefix(remoteURL, "git+") && strings.Contains(remoteURL, "://") {
		remoteURL = strings.TrimPrefix(remoteURL, "git+")
	}

	if !strings.Contains(remoteURL, "://") {
		remoteURL = strings.Replace(remoteURL, ":", "/", 1)
		remoteURL = "https://" + remoteURL
	}

	remoteURL = fixGitOrSSHRepoURL(remoteURL)

	u, err := url.Parse(remoteURL)

	if err != nil {
		return "", err
	}

	u.User = nil

	return u.String(), nil
}

func fixGitOrSSHRepoURL(remoteURL string) string {
	var prefixes = []string{"git+https://", "git://", "ssh://", "git+ssh://"}

	for _, p := range prefixes {
		if !strings.HasPrefix(remoteURL, p) {
			continue
		}

		remoteURL = strings.TrimPrefix(remoteURL, p)
		remoteURL = "https://" + strings.Replace(remoteURL, ":", "/", 1)
		break
	}

	return remoteURL
}
