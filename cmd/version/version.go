package cmdversion

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/verbose"
)

// VersionCmd is used for reading the version of this tool
var VersionCmd = &cobra.Command{
	Use:   "version",
	Run:   versionRun,
	Short: "Show CLI version",
}

func versionRun(cmd *cobra.Command, args []string) {
	var os = runtime.GOOS
	var arch = runtime.GOARCH
	fmt.Printf(
		"WeDeploy CLI version %s %s/%s\n",
		defaults.Version,
		os,
		arch)

	if defaults.Build != "" {
		fmt.Printf("Build commit: %v\n", defaults.Build)
	}

	if defaults.BuildTime != "" {
		fmt.Printf("Build time: %v\n", defaults.BuildTime)
	}

	verbose.Debug("Go version: %v\n", runtime.Version())
}
