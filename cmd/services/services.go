package services

import (
	"github.com/launchpad-project/cli/services"
	"github.com/spf13/cobra"
)

// PodServicesCmd is used for managing the pod services
var PodServicesCmd = &cobra.Command{
	Use:   "pod",
	Short: "List pods",
	Run:   podServicesRun,
}

// ServicesCmd is used for managing the services
var ServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "Manage Launchpad services",
}

func init() {
	ServicesCmd.AddCommand(PodServicesCmd)
}

func podServicesRun(cmd *cobra.Command, args []string) {
	services.GetPods()
}
