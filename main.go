/*
cli.cmd

	https://github.com/wedeploy/cli

*/

package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/wedeploy/cli/cmd"
	"github.com/wedeploy/cli/cmd/gitcredentialhelper"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	var args = os.Args

	if len(args) >= 2 && args[1] == "git-credential-helper" {
		var err = gitcredentialhelper.Run(args)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		return
	}

	cmd.Execute()
}
