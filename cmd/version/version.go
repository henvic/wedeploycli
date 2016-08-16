package cmdversion

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/defaults"
)

// VersionCmd is used for reading the version of this tool
var VersionCmd = &cobra.Command{
	Use:   "version",
	Run:   versionRun,
	Short: "Print version information and quit",
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
		fmt.Printf("build %v\n", defaults.Build)
	}
}
