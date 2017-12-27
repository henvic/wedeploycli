package legal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/legal"
	"github.com/wedeploy/cli/verbose"
)

// LegalCmd is used for showing the abouts of all used libraries
var LegalCmd = &cobra.Command{
	Use:   "legal",
	RunE:  legalRun,
	Args:  cobra.NoArgs,
	Short: "Legal notices",
}

func legalRun(cmd *cobra.Command, args []string) error {
	pager := exec.Command("less")
	pager.Stdin = strings.NewReader(getLegalNotices())
	pager.Stdout = os.Stdout
	err := pager.Run()

	if err != nil {
		verbose.Debug("can't page with less: " + err.Error())
		fmt.Print(getLegalNotices())
	}

	return nil
}

func getLegalNotices() string {
	return fmt.Sprintf(`Legal Notices:

Copyright © 2016-present Liferay, Inc.
WeDeploy CLI Software License Agreement

Liferay, the Liferay logo, WeDeploy, and WeDeploy logo
are trademarks of Liferay, Inc., registered in the U.S. and other countries.

Acknowledgements:
Portions of this Liferay software may utilize the following copyrighted material,
the use of which is hereby acknowledged.

%s`, legal.Licenses)
}
