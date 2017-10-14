package cmdcheck

import (
	"context"
	"fmt"
	"os"
	"time"

	humanize "github.com/dustin/go-humanize"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/diagnostics"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/verbose"
)

var (
	serial  bool
	print   bool
	send    bool
	timeout = 15
)

// DiagnosticsCmd sets the user credential
var DiagnosticsCmd = &cobra.Command{
	Use:     "diagnostics",
	Short:   "Run system diagnostics and show report",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    diagnosticsRun,
	Aliases: []string{"check"},
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
		diagnostics.Write(os.Stderr, report)
	}

	fmt.Println(fancy.Info("Diagnostics report size: ") +
		color.Format(color.Bold, humanize.Bytes(uint64(report.Len()))))

	if !send && !cmd.Flag("send").Changed {
		var report, askErr = fancy.Boolean("Send this report to WeDeploy?")

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

	cloudRemote, ok := wectx.Config().Remotes[defaults.CloudRemote]

	if ok {
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
		"Send to WeDeploy")

	DiagnosticsCmd.Flags().IntVar(
		&timeout,
		"timeout",
		15,
		"Timeout for the diagnostics in seconds")

	DiagnosticsCmd.Flag("serial").Hidden = true
	DiagnosticsCmd.Flag("timeout").Hidden = true
}
