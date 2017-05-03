package cmddiagnostics

import (
	"context"
	"fmt"
	"os"
	"time"

	"strings"

	"github.com/hashicorp/errwrap"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"github.com/wedeploy/cli/cmdargslen"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/diagnostics"
	"github.com/wedeploy/cli/prompt"
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
	Short:   "Diagnose & Feedback",
	PreRunE: cmdargslen.ValidateCmd(0, 0),
	RunE:    diagnosticsRun,
	Hidden:  true,
}

func diagnosticsRun(cmd *cobra.Command, args []string) error {
	var d = diagnostics.Diagnostics{
		Timeout:     time.Duration(timeout) * time.Second,
		Executables: diagnostics.Executables,
		Serial:      serial,
	}

	fmt.Println("Running diagnostics tools…")

	var ctx, cancel = d.Start()
	<-ctx.Done()
	cancel()

	var report = d.Collect()

	if print {
		diagnostics.Write(os.Stderr, report)
	}

	if !send && !cmd.Flag("send").Changed {
		const question = "Press [Enter] or type \"yes\" send diagnostics report to WeDeploy [yes/no]"
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

	if config.Global != nil {
		username = config.Context.Username
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
