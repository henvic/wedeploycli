/*
cli.cmd

	https://github.com/wedeploy/cli

*/

package main

import (
	"math/rand"
	"time"

	"github.com/wedeploy/cli/cmd"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	cmd.Execute()
}
