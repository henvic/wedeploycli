package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/launchpad-project/cli/cmd/auth"
	cmdcontainers "github.com/launchpad-project/cli/cmd/containers"
	cmdcreate "github.com/launchpad-project/cli/cmd/createctx"
	cmddeploy "github.com/launchpad-project/cli/cmd/deploy"
	cmdhooks "github.com/launchpad-project/cli/cmd/hooks"
	cmdlogs "github.com/launchpad-project/cli/cmd/logs"
	cmdprojects "github.com/launchpad-project/cli/cmd/projects"
	cmdrestart "github.com/launchpad-project/cli/cmd/restart"
	"github.com/launchpad-project/cli/cmd/run"
	cmdupdate "github.com/launchpad-project/cli/cmd/update"
	cmdversion "github.com/launchpad-project/cli/cmd/version"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/configstore"
	"github.com/launchpad-project/cli/defaults"
	"github.com/launchpad-project/cli/update"
	"github.com/launchpad-project/cli/verbose"
	"github.com/spf13/cobra"
)

// WhitelistCmdsNoAuthentication for cmds that doesn't require authentication
var WhitelistCmdsNoAuthentication = map[string]bool{
	"login":   true,
	"logout":  true,
	"build":   true,
	"deploy":  true,
	"update":  true,
	"version": true,
}

// RootCmd is the main command for the CLI
var RootCmd = &cobra.Command{
	Use:   "we",
	Short: "WeDeploy CLI tool",
	Long: `WeDeploy Command Line Interface
Version ` + defaults.Version + `
Copyright 2016 Liferay, Inc.
http://liferay.io`,
	PersistentPreRun: persistentPreRun,
	Run:              run,
}

var version bool
var globalStore *configstore.Store

// Execute is the Entry-point for the CLI
func Execute() {
	var wgUpdate sync.WaitGroup
	var errUpdate error

	wgUpdate.Add(1)
	go func() {
		errUpdate = update.NotifierCheck()
		wgUpdate.Done()
	}()

	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}

	wgUpdate.Wait()

	if errUpdate == nil {
		update.Notify()
	} else {
		println("Update notification error:", errUpdate.Error())
	}
}

var commands = []*cobra.Command{
	cmdauth.LoginCmd,
	cmdauth.LogoutCmd,
	cmdcreate.CreateCmd,
	cmdlogs.LogsCmd,
	cmdprojects.ProjectsCmd,
	cmdcontainers.ContainersCmd,
	cmdrestart.RestartCmd,
	cmdhooks.BuildCmd,
	cmdrun.RunCmd,
	cmddeploy.DeployCmd,
	cmdupdate.UpdateCmd,
	cmdversion.VersionCmd,
}

func init() {
	config.Setup()

	RootCmd.PersistentFlags().BoolVarP(
		&verbose.Enabled,
		"verbose",
		"v",
		false,
		"Verbose output")

	RootCmd.PersistentFlags().BoolVar(
		&color.NoColor,
		"no-color",
		false,
		"Disable color output")

	RootCmd.Flags().BoolVar(
		&version,
		"version", false, "Print version information and quit")
	RootCmd.Flags().MarkHidden("version")

	var csg = config.Stores["global"]

	if csg.Get("no_color") == "true" {
		color.NoColor = true
	}

	for _, c := range commands {
		RootCmd.AddCommand(c)
	}
}

func persistentPreRun(cmd *cobra.Command, args []string) {
	verifyAuth(cmd.CommandPath())
}

func run(cmd *cobra.Command, args []string) {
	if version {
		cmdversion.VersionCmd.Run(cmd, args)
	} else {
		cmd.Help()
	}
}

func isCmdWhitelistNoAuth(commandPath string) bool {
	var parts = strings.SplitAfterN(commandPath, " ", 2)

	if len(parts) < 2 {
		return true
	}

	for key := range WhitelistCmdsNoAuthentication {
		if key == parts[1] {
			return true
		}
	}

	return false
}

func verifyAuth(commandPath string) {
	var csg = config.Stores["global"]

	if isCmdWhitelistNoAuth(commandPath) {
		return
	}

	_, err1 := csg.GetString("endpoint")
	_, err2 := csg.GetString("username")
	_, err3 := csg.GetString("password")

	if err1 == nil && err2 == nil && err3 == nil {
		return
	}

	pleaseLoginFeedback()
}

func pleaseLoginFeedback() {
	fmt.Fprintf(os.Stderr, "Please run \"we login\" first.\n")
	os.Exit(1)
}
