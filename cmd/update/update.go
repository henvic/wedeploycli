package cmdupdate

import (
	"fmt"

	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/update"
	"github.com/spf13/cobra"
)

// UpdateCmd is used for updating this tool
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Run:   updateRun,
	Short: "Updates this tool to the latest version",
}

func updateRun(cmd *cobra.Command, args []string) {
	fmt.Println("Trying to update Launchpad CLI")
	fmt.Println("Current installed version is " + launchpad.Version)
	update.ToLatest()
}
