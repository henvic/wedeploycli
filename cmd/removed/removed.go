package cmdremoved

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/color"
)

// List of removed commands that must be dealt with
var List = []*cobra.Command{
	createCmd,
	runCmd,
	stopCmd,
	linkCmd,
	unlinkCmd,
}

var createCmd = &cobra.Command{
	Use:                "create",
	Hidden:             true,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr,
			"Use instead: "+color.Format(color.FgHiRed, "we generate %v", strings.Join(args, " ")))
		os.Exit(1)
	},
}

var runCmd = &cobra.Command{
	Use:                "run",
	Hidden:             true,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr,
			color.Format(color.FgHiRed, "we run")+
				" is no longer required to try WeDeploy locally.")
		os.Exit(1)
	},
}

var stopCmd = &cobra.Command{
	Use:                "stop",
	Hidden:             true,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr,
			color.Format(color.FgHiRed, "we stop")+
				" has been replaced by "+
				color.Format(color.FgHiRed, "we dev --shutdown-infra")+".")
		os.Exit(1)
	},
}

var linkCmd = &cobra.Command{
	Use:                "link",
	Hidden:             true,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr,
			"Use instead: "+color.Format(color.FgHiRed, "we dev %v", strings.Join(args, " ")))

		os.Exit(1)
	},
}

var unlinkCmd = &cobra.Command{
	Use:                "unlink",
	Hidden:             true,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(os.Stderr,
			"Use instead: "+color.Format(color.FgHiRed, "we dev stop %v", strings.Join(args, " ")))
		os.Exit(1)
	},
}
