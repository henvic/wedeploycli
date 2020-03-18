package metrics

import (
	"context"
	"fmt"

	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/metrics"
	"github.com/spf13/cobra"
)

var noPurge bool

func init() {
	MetricsCmd.AddCommand(UsageCmd)
	UsageCmd.AddCommand(submitCmd)

	MetricsCmd.Hidden = true

	submitCmd.Flags().BoolVar(
		&noPurge,
		"no-purge",
		false,
		"Do not purge usage log after submission")
}

// MetricsCmd controls the metrics for the Liferay Cloud Platform CLI
var MetricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Metrics",
}

// UsageCmd unsets the user credential
var UsageCmd = &cobra.Command{
	Use:   "usage",
	Short: "CLI usage submission tool",
}

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit anonymous analytics to Liferay Cloud",
	Long: `Submit anonymous analytics to Liferay Cloud

Analytics events are stashed in ~/.lcp_metrics and occasionally
submitted to Liferay Cloud by this command. No need to call it yourself.`,
	RunE: submitRun,
}

func submitRun(cmd *cobra.Command, args []string) error {
	fmt.Println(metrics.PreparingMetricsText)

	var sender = &metrics.Sender{
		Purge: !noPurge,
	}

	wectx := we.Context()
	var events, err = sender.TrySubmit(context.Background(), wectx.Config())

	if err != nil {
		return err
	}

	fmt.Printf("%v lines of events sent.\n", events)
	return nil
}
