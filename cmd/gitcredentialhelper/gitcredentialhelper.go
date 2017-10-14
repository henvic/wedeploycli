package gitcredentialhelper

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/envs"
)

// GitCredentialHelperCmd is used for reading the version of this tool
var GitCredentialHelperCmd = &cobra.Command{
	Use:    "git-credential-helper",
	RunE:   run,
	Hidden: true,
	Short:  "Git credential helper for WeDeploy",
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("usage: we git-credential-helper get")
	}

	// this is a read-only credential helper: ignore transparently other commands
	// https://www.kernel.org/pub/software/scm/git/docs/technical/api-credentials.html
	if args[0] != "get" {
		return nil
	}

	var token = os.Getenv(envs.GitCredentialRemoteToken)

	if token == "" {
		return errors.New("internal command: missing credentials")
	}

	fmt.Fprintf(os.Stdout, "username=%s\npassword=\n", token)
	return nil
}
