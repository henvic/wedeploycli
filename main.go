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
	"github.com/wedeploy/cli/envs"
)

func maybeSetCustomTimezone() {
	timezone := os.Getenv(envs.TZ)

	if timezone == "" {
		return
	}

	l, err := time.LoadLocation(timezone)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failure setting a custom timezone: %+v\n", err)
		return
	}

	time.Local = l
}

func main() {
	maybeSetCustomTimezone()

	rand.Seed(time.Now().UTC().UnixNano())

	var args = os.Args

	if len(args) >= 2 && args[1] == "git-credential-helper" {
		var err = gitcredentialhelper.Run(args)

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		return
	}

	cmd.Execute()
}
