package diagnostics

import (
	"context"
	"fmt"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/command/internal/we"
	"github.com/henvic/wedeploycli/defaults"
	"github.com/henvic/wedeploycli/diagnostics"
	"github.com/henvic/wedeploycli/fancy"
	"github.com/henvic/wedeploycli/verbose"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
)

var (
	serial  bool
	print   bool
	send    bool
	timeout = 15
)

// DiagnosticsCmd sets the user credential
var DiagnosticsCmd = &cobra.Command{
	Use:    "diagnostics",
	Short:  "Run system diagnostics and show report",
	RunE:   diagnosticsRun,
	Args:   cobra.NoArgs,
	Hidden: true,
}

func diagnosticsRun(cmd *cobra.Command, args []string) error {
	print = print || verbose.Enabled
	var d = diagnostics.Diagnostics{
		Timeout:     time.Duration(timeout) * time.Second,
		Executables: diagnostics.Executables,
		Serial:      serial,
	}

	fmt.Println("Running diagnostics tools...")
	d.Run(context.Background())

	var report = d.Collect()
	fmt.Println()

	if print {
		fmt.Printf("%s", report)
	}

	bu := uint64(len([]byte(report)))

	fmt.Println(fancy.Info("Diagnostics report size: ") +
		color.Format(color.Bold, humanize.Bytes(bu)))

	if !send && !cmd.Flag("send").Changed {
		var report, askErr = fancy.Boolean("Send this report to Liferay Cloud?")

		if askErr != nil {
			return askErr
		}

		if report {
			send = true
		}
	}

	if !send {
		return nil
	}

	return submit(report)
}

func submit(report diagnostics.Report) error {
	var username string
	var wectx = we.Context()
	var conf = wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes

	if rl.Has(defaults.CloudRemote) {
		cloudRemote := rl.Get(defaults.CloudRemote)
		username = cloudRemote.Username
	}

	var entry = diagnostics.Entry{
		ID:       uuid.NewV4().String(),
		Username: username,
		Report:   report,
	}

	var err = diagnostics.Submit(context.Background(), entry)

	if err != nil {
		return err
	}

	fmt.Println(fancy.Info("Report ID: ") + entry.ID)
	fmt.Println(fancy.Info("In case you need support, providing the ID will help us to diagnose your situation."))
	return nil
}

func init() {
	DiagnosticsCmd.Flags().BoolVar(
		&serial,
		"serial",
		false,
		"Do not run diagnostics in parallel")

	DiagnosticsCmd.Flags().BoolVar(
		&print,
		"print",
		false,
		"Print diagnostics")

	DiagnosticsCmd.Flags().BoolVar(
		&send,
		"send",
		false,
		"Send to Liferay Cloud")

	DiagnosticsCmd.Flags().IntVar(
		&timeout,
		"timeout",
		15,
		"Timeout for the diagnostics in seconds")

	DiagnosticsCmd.Flag("serial").Hidden = true
	DiagnosticsCmd.Flag("timeout").Hidden = true
}
