package cmddiagnostics

import (
	"context"
	"fmt"
	"os"
	"time"

	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/hashicorp/errwrap"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/diagnostics"
	"github.com/wedeploy/cli/prompt"
	"github.com/wedeploy/cli/run"
	"github.com/wedeploy/cli/verbose"
)

var (
	dryRun  bool
	serial  bool
	print   bool
	send    bool
	timeout = 15
)

// DiagnosticsCmd sets the user credential
var DiagnosticsCmd = &cobra.Command{
	Use:     "diagnostics",
	Short:   "Troubleshoot problems and submit system diagnostics\n",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    diagnosticsRun,
}

func diagnosticsRun(cmd *cobra.Command, args []string) error {
	print = print || verbose.Enabled
	var d = diagnostics.Diagnostics{
		Timeout:     time.Duration(timeout) * time.Second,
		Executables: diagnostics.Executables,
		Serial:      serial,
	}

	fmt.Println("Running diagnostics tools...")
	run.MaybeStartDocker()
	d.Run(context.Background())

	var report = d.Collect()
	fmt.Println()

	if print {
		diagnostics.Write(os.Stderr, report)
	}

	fmt.Printf("Diagnostics report size: %v\n", color.Format(color.Bold, color.FgHiBlue, humanize.Bytes(uint64(report.Len()))))

	if !send && !cmd.Flag("send").Changed {
		const question = "Press [Enter] or type \"yes\" to send diagnostics report to WeDeploy [yes/no]"
		switch ask, askErr := prompt.Prompt(question); {
		case askErr != nil:
			return errwrap.Wrapf("Skipping diagnostics submission: {{err}}", askErr)
		case len(ask) == 0 || strings.ToLower(ask[:1]) == "y":
			send = true
		}
	}

	if !send {
		return nil
	}

	return submit(report)
}

func printUsername() string {
	if config.Global == nil || config.Context.Username == "" {
		return "not logged in"
	}

	return config.Context.Username
}

func submit(report diagnostics.Report) error {
	var username string

	cloudRemote, ok := config.Global.Remotes[defaults.CloudRemote]

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

	if username != "" {
		fmt.Println("Username: " + username)
	}

	fmt.Println("Diagnostics ID: " + entry.ID)
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
		"send to WeDeploy")

	DiagnosticsCmd.Flags().IntVar(
		&timeout,
		"timeout",
		15,
		"Timeout for the diagnostics in seconds")

	DiagnosticsCmd.Flag("serial").Hidden = true
	DiagnosticsCmd.Flag("timeout").Hidden = true
}
