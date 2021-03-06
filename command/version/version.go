package version

import (
	"fmt"
	"runtime"

	"github.com/henvic/wedeploycli/defaults"
	"github.com/henvic/wedeploycli/verbose"
	"github.com/spf13/cobra"
)

// VersionCmd is used for reading the version of this tool
var VersionCmd = &cobra.Command{
	Use:   "version",
	Args:  cobra.NoArgs,
	Run:   versionRun,
	Short: "Show CLI version",
}

// Print the current version
func Print() {
	var os = runtime.GOOS
	var arch = runtime.GOARCH
	fmt.Printf(
		"Liferay Cloud Platform CLI version %s %s/%s\n",
		defaults.Version,
		os,
		arch)

	if defaults.Build != "" {
		fmt.Printf("Build commit: %v\n", defaults.Build)
	}

	if defaults.BuildTime != "" {
		fmt.Printf("Build time: %v\n", defaults.BuildTime)
	}

	verbose.Debug("Go version:", runtime.Version())
}

func versionRun(cmd *cobra.Command, args []string) {
	Print()
}
