/*
api.cmd

	https://github.com/wedeploy/cli

*/

package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/autocomplete"
	"github.com/wedeploy/cli/cmd"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/color/template"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/flagsfromhost"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/login"
	"github.com/wedeploy/cli/metrics"
	"github.com/wedeploy/cli/update"
	"github.com/wedeploy/cli/userhome"
	"github.com/wedeploy/cli/verbose"
)

// Windows users using Prompt should see no color
// Issue #51.
// https://github.com/wedeploy/cli/issues/51
func turnColorsOffOnWindows() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	_, windowsPrompt := os.LookupEnv("PROMPT")
	return windowsPrompt
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	cobra.AddTemplateFuncs(colortemplate.Functions())

	_, machineFriendly := os.LookupEnv("WEDEPLOY_MACHINE_FRIENDLY")
	formatter.Human = !machineFriendly

	if isCommand("--no-color") || turnColorsOffOnWindows() {
		color.NoColorFlag = true
	}
}

func main() {
	var panickingFlag = true
	defer verbosePrinting()
	defer panickingListener(&panickingFlag)

	setErrorHandlingCommandName()
	(&mainProgram{}).run()
	panickingFlag = false
}

type mainProgram struct {
	cmd            *cobra.Command
	cmdErr         error
	cmdFriendlyErr error
}

func (m *mainProgram) run() {
	m.setupMetrics()
	(&configLoader{}).load()
	var uc update.Checker

	if !isCommand("autocomplete") && !isCommand("metrics") && !isCommand("build") {
		uc.Check()
	}

	m.executeCommand()
	uc.Feedback()

	autocomplete.AutoInstall()
	m.maybeSubmitAnalyticsReport()
}

func (m *mainProgram) setupMetrics() {
	metrics.SetPID(os.Getpid())
	metrics.SetPath(filepath.Join(userhome.GetHomeDir(), ".we_metrics"))
}

func printError(e error) {
	fmt.Fprintf(os.Stderr, "%v\n", fancy.Error(e))

	var aft = errwrap.GetType(e, &apihelper.APIFault{})

	if aft == nil {
		return
	}

	af, ok := aft.(*apihelper.APIFault)

	if !ok || af.Status < 500 || af.Status > 599 {
		return
	}

	fmt.Fprintf(os.Stderr, "%v\n",
		fancy.Error("Contact us: "+defaults.SupportEmail))
}

func hideHelpFlag() {
	// hide the --help flag on all commands, but top-level
	cmd.RootCmd.PersistentFlags().BoolP("help", "h", false, "Show help message")
	_ = cmd.RootCmd.PersistentFlags().MarkHidden("help")
	cmd.RootCmd.Flags().BoolP("help", "h", false, "Show help message")
}

func (m *mainProgram) executeCommand() {
	hideHelpFlag()
	m.cmd, m.cmdErr = cmd.RootCmd.ExecuteC()
	m.cmdFriendlyErr = errorhandling.Handle(m.cmdErr)

	switch m.cmdErr.(type) {
	case login.CanceledCommand:
		fmt.Fprintln(os.Stderr, fancy.Success(m.cmdErr))
		m.cmdErr = nil
	}

	if m.cmdErr != nil {
		fmt.Fprintln(os.Stderr, fancy.Error(
			fmt.Sprintf(`Something went wrong with your "%s" operation.`, m.cmd.UseLine())))
		printError(m.cmdFriendlyErr)
	}

	m.reportCommand()

	if m.cmdErr != nil {
		m.commandErrorConditionalUsage()
		errorhandling.RunAfterError()
		os.Exit(1)
	}
}

func (m *mainProgram) getCommandFlags() []string {
	var flags = []string{}

	m.cmd.Flags().Visit(func(f *pflag.Flag) {
		flags = append(flags, f.Name)
	})

	return flags
}

func (m *mainProgram) getCommandErrorDetails() map[string]string {
	if m.cmdErr == nil {
		return nil
	}

	var extra = map[string]string{
		"error_type": errorhandling.GetTypes(m.cmdErr),
	}

	if m.cmdErr.Error() != m.cmdFriendlyErr.Error() {
		// currently (?; as of 3rd, Nov 2016) the friendly error message list is
		// static; this might need to be improved later if the list starts accepting
		// values like it was a template system
		extra["friendly_error"] = m.cmdFriendlyErr.Error()
	}

	return extra
}

func (m *mainProgram) reportCommand() {
	var commandPath = m.cmd.CommandPath()

	if commandPath == "we metrics usage reset" {
		// Skip storing "we metrics usage reset" on the analytics log
		// otherwise this would recreate the file just after removal
		return
	}

	metrics.Rec(metrics.Event{
		Type:  "command_exec",
		Text:  commandPath,
		Tags:  m.getCommandFlags(),
		Extra: m.getCommandErrorDetails(),
	})
}

func (m *mainProgram) maybeSubmitAnalyticsReport() {
	if !isCommand("metrics") {
		if err := metrics.SubmitEventuallyOnBackground(); err != nil {
			fmt.Fprintf(os.Stderr,
				"Error trying to submit analytics on background: %v\n",
				errorhandling.Handle(err))
		}
	}
}

type configLoader struct {
	reload bool
}

func (cl *configLoader) loadConfig() {
	var path = filepath.Join(userhome.GetHomeDir(), ".we")
	if err := config.Setup(path); err != nil {
		printError(errorhandling.Handle(err))
		os.Exit(1)
	}

	if config.Global.NoColor {
		color.NoColor = true
	}

	flagsfromhost.InjectRemotes(&(config.Global.Remotes))
}

func (cl *configLoader) checkPastVersion() {
	if config.Global.PastVersion != "" {
		update.ApplyTransitions(config.Global.PastVersion)
		config.Global.PastVersion = ""
		cl.reload = true
	}
}

func (cl *configLoader) checkAnalytics() {
	if !config.Global.EnableAnalytics || config.Global.AnalyticsID != "" {
		return
	}

	if err := metrics.Enable(); err != nil {
		verbose.Debug("Error trying to fix enabled metrics without analytics ID: " + err.Error())
	}
}

func (cl *configLoader) applyChanges() {
	if !cl.reload {
		return
	}

	if err := config.Global.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", errorhandling.Handle(err))
		os.Exit(1)
	}

	cl.loadConfig()
}

func (cl *configLoader) load() {
	cl.loadConfig()
	cl.checkAnalytics()
	cl.checkPastVersion()
	cl.applyChanges()
}

func isCommand(cmd string) bool {
	for _, s := range os.Args {
		if s == cmd {
			return true
		}
	}

	return false
}

func setErrorHandlingCommandName() {
	errorhandling.CommandName = strings.Join(os.Args[1:], " ")
}

func verbosePrinting() {
	if verbose.Defered {
		verbose.PrintDefered()
	}
}

func panickingListener(panicking *bool) {
	if !*panicking {
		return
	}

	errorhandling.Info()
	// don't recover from panic to get more context
	// to avoid having to handle it
	// unless we find out it is really useful later
	metrics.Rec(metrics.Event{
		Type: "panic",
	})
}

func (m *mainProgram) commandErrorConditionalUsage() {
	// this tries to print the usage for a given command only when one of the
	// errors below is caused by cobra
	var emsg = m.cmdErr.Error()
	if strings.HasPrefix(emsg, "unknown flag: ") ||
		strings.HasPrefix(emsg, "unknown shorthand flag: ") ||
		strings.HasPrefix(emsg, "invalid argument ") ||
		strings.HasPrefix(emsg, "bad flag syntax: ") ||
		strings.HasPrefix(emsg, "flag needs an argument: ") {
		if ue := m.cmd.Usage(); ue != nil {
			panic(ue)
		}
	} else if strings.HasPrefix(emsg, "unknown command ") {
		println(fancy.Error("Run 'we --help' for usage."))
	}
}
