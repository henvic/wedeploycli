package gitcredentialhelper

import (
	"errors"
	"fmt"
	"os"

	"github.com/henvic/wedeploycli/envs"
)

// Run the credential helper
func Run(args []string) error {
	if len(args) != 3 {
		return errors.New("usage: lcp git-credential-helper get")
	}

	// this is a read-only credential helper: ignore transparently other commands
	// https://www.kernel.org/pub/software/scm/git/docs/technical/api-credentials.html
	if args[2] != "get" {
		return nil
	}

	var token = os.Getenv(envs.GitCredentialRemoteToken)

	if token == "" {
		return errors.New("internal command: missing credentials")
	}

	fmt.Printf("username=%s\npassword=\n", token)
	return nil
}
