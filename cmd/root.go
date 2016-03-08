package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/cmd/auth"
	cconfig "github.com/launchpad-project/cli/cmd/config"
	"github.com/launchpad-project/cli/cmd/hooks"
	"github.com/launchpad-project/cli/cmd/info"
	"github.com/launchpad-project/cli/cmd/update"
	"github.com/launchpad-project/cli/cmd/version"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/configstore"
	"github.com/launchpad-project/cli/verbose"
	"github.com/spf13/cobra"
)

// WhitelistCmdsNoAuthentication for cmds that doesn't require authentication
var WhitelistCmdsNoAuthentication = map[string]bool{
	"login":   true,
	"logout":  true,
	"info":    true,
	"config":  true,
	"update":  true,
	"version": true,
}

// RootCmd is the main command for the CLI
var RootCmd = &cobra.Command{
	Use:   "launchpad",
	Short: "Launchpad CLI tool",
	Long: `Launchpad Command Line Interface
Version ` + launchpad.Version + `
Copyright 2016 Liferay, Inc.
http://liferay.io`,
	PersistentPreRun: persistentPreRun,
}

var globalStore *configstore.Store

// Execute is the Entry-point for the CLI
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	config.Setup()

	RootCmd.PersistentFlags().BoolVarP(
		&verbose.Enabled,
		"verbose",
		"v",
		false,
		"verbose output")

	RootCmd.AddCommand(auth.LoginCmd)
	RootCmd.AddCommand(auth.LogoutCmd)
	RootCmd.AddCommand(cconfig.ConfigCmd)
	RootCmd.AddCommand(update.UpdateCmd)
	RootCmd.AddCommand(info.InfoCmd)
	RootCmd.AddCommand(hooks.BuildCmd)
	RootCmd.AddCommand(hooks.DeployCmd)
	RootCmd.AddCommand(version.VersionCmd)
}

func persistentPreRun(cmd *cobra.Command, args []string) {
	verifyAuth(cmd.CommandPath())
}

func verifyAuth(commandPath string) {
	var csg = config.Stores["global"]
	var test = strings.SplitAfterN(commandPath, " ", 2)[1]

	for key := range WhitelistCmdsNoAuthentication {
		if key == test {
			return
		}
	}

	_, err1 := csg.GetString("endpoint")
	_, err2 := csg.GetString("username")
	_, err3 := csg.GetString("password")

	if err1 == nil && err2 == nil && err3 == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "Please run \"launchpad login\" first.\n")
	os.Exit(1)
}
