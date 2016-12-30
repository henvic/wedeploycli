package cmdmetrics

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/metrics"
)

var noPurge bool

func init() {
	MetricsCmd.AddCommand(UsageCmd)
	UsageCmd.AddCommand(statusCmd)
	UsageCmd.AddCommand(enableCmd)
	UsageCmd.AddCommand(disableCmd)
	UsageCmd.AddCommand(resetCmd)
	UsageCmd.AddCommand(purgeCmd)
	UsageCmd.AddCommand(submitCmd)

	submitCmd.Flags().BoolVar(
		&noPurge,
		"no-purge",
		false,
		"Do not purge usage log after submission")
}

// MetricsCmd controls the metrics for the WeDeploy CLI
var MetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Analytics reporting and settings",
}

// UsageCmd unsets the user credential
var UsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Usage preferences and submit tool",
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get current status [enabled | disabled]",
	RunE:  statusRun,
}

func statusRun(cmd *cobra.Command, args []string) error {
	switch config.Global.EnableAnalytics {
	case true:
		fmt.Println("enabled")
	default:
		fmt.Println("disabled")
	}

	return nil
}

var enableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable usage reporting to WeDeploy",
	RunE:  enableRun,
}

func enableRun(cmd *cobra.Command, args []string) error {
	return metrics.Enable()
}

var disableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable usage reporting to WeDeploy",
	RunE:  disableRun,
}

func disableRun(cmd *cobra.Command, args []string) error {
	return metrics.Disable()
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Purge existing analytics usage file and reset usage log session ID (SID)",
	RunE:  resetRun,
}

func resetRun(cmd *cobra.Command, args []string) error {
	return metrics.Reset()
}

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "purge analytics file (~/.we_metrics) contents",
	RunE:  purgeRun,
}

func purgeRun(cmd *cobra.Command, args []string) error {
	return metrics.Purge()
}

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit anonymous analytics to WeDeploy",
	Long: `Submit anonymous analytics to WeDeploy

Analytics events are stashed in ~/.we_metrics and occasionally
submitted to WeDeploy by this command. No need to call it yourself.`,
	RunE: submitRun,
}

func submitRun(cmd *cobra.Command, args []string) error {
	var sender = &metrics.Sender{
		Purge: !noPurge,
	}

	var events, err = sender.TrySubmit()

	if err != nil {
		return err
	}

	fmt.Printf("%v lines of events sent.\n", events)
	return nil
}
