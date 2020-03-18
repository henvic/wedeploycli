package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/autocomplete"
	"github.com/wedeploy/cli/color"
	colortemplate "github.com/wedeploy/cli/color/template"
	"github.com/wedeploy/cli/command/canceled"
	cmdexec "github.com/wedeploy/cli/command/exec"
	execargs "github.com/wedeploy/cli/command/exec/args"
	"github.com/wedeploy/cli/command/internal/we"
	"github.com/wedeploy/cli/command/root"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/envs"
	"github.com/wedeploy/cli/errorhandler"
	"github.com/wedeploy/cli/exiterror"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/links"
	"github.com/wedeploy/cli/metrics"
	"github.com/wedeploy/cli/shell"
	"github.com/wedeploy/cli/update"
	"github.com/wedeploy/cli/userhome"
	"github.com/wedeploy/cli/verbose"
)

// Execute runs the application
func Execute() {
	var panickingFlag = true
	defer verbose.PrintDeferred()
	defer panickingListener(&panickingFlag)

	setErrorHandlingCommandName()
	(&mainProgram{}).run()
	panickingFlag = false
}

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
	cobra.AddTemplateFuncs(colortemplate.Functions())

	_, machineFriendly := os.LookupEnv(envs.MachineFriendly)
	formatter.Human = !machineFriendly

	if isCommand("--no-color") || turnColorsOffOnWindows() {
		color.NoColorFlag = true
		color.NoColor = true
	}
}

type mainProgram struct {
	cmd            *cobra.Command
	cmdErr         error
	cmdFriendlyErr error
	config         *config.Config
}

func (m *mainProgram) run() {
	m.setupMetrics()
	m.config = (&configLoader{}).load()

	var uc update.Checker

	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	if !isCommand("autocomplete") && !isCommand("metrics") {
		uc.Check(ctx, m.config)
	}

	m.prepareCommand()
	m.executeCommand()
	cancel()

	uc.Feedback(m.config)
	m.autocomplete()
	m.maybeSubmitAnalyticsReport()
}

func (m *mainProgram) autocomplete() {
	var conf = m.config
	var params = conf.GetParams()

	if !params.NoAutocomplete {
		autocomplete.AutoInstall()
	}
}

func (m *mainProgram) setupMetrics() {
	metrics.SetPID(os.Getpid())

	var weMetricsPath = filepath.Join(userhome.GetHomeDir(), ".lcp_metrics")
	metrics.SetPath(weMetricsPath)
}

func printError(e error) {
	_, _ = fmt.Fprintf(os.Stderr, "%v\n", fancy.Error(e))

	var aft = errwrap.GetType(e, apihelper.APIFault{})

	if aft == nil {
		return
	}

	af, ok := aft.(apihelper.APIFault)

	if !ok || af.Status < 500 || af.Status > 599 {
		return
	}

	_, _ = fmt.Fprintf(os.Stderr, "%v\n",
		fancy.Error("Contact us: "+links.Support))
}

func (m *mainProgram) prepareCommand() {
	if isCommand("exec") {
		if execArgs, rewrite := execargs.MaybeRewrite(cmdexec.ExecCmd, os.Args[1:]); rewrite {
			root.Cmd.SetArgs(execArgs)
		}
	}
}

func (m *mainProgram) executeCommand() {
	m.cmd, m.cmdErr = root.Cmd.ExecuteC()
	m.cmdFriendlyErr = errorhandler.Handle(m.cmdErr)

	m.maybePrintCommandError()
	m.reportCommand()

	if m.cmdErr != nil {
		m.commandErrorConditionalUsage()
		m.handleErrorExitCode()
	}
}

func (m *mainProgram) maybePrintCommandError() {
	// maybe should use errwrap.GetType instead
	switch m.cmdErr.(type) {
	case canceled.Command:
		cc := m.cmdErr.(canceled.Command)

		if !cc.Quiet() {
			printError(m.cmdErr)
		}

		m.cmdErr = nil
	case *exec.ExitError, *shell.ExitError: // don't print error message
	default:
		if m.cmdErr != nil {
			printError(m.cmdFriendlyErr)
		}
	}
}

func (m *mainProgram) handleErrorExitCode() {
	errorhandler.RunAfterError()
	verbose.PrintDeferred()

	var maybeExitErr = errwrap.GetType(m.cmdErr, &exec.ExitError{})

	if ee, ok := maybeExitErr.(*exec.ExitError); ok {
		ws := ee.Sys().(syscall.WaitStatus)
		os.Exit(ws.ExitStatus())
	}

	var maybeSSHExitErr = errwrap.GetType(m.cmdErr, &shell.ExitError{})

	if sshE, ok := maybeSSHExitErr.(*shell.ExitError); ok {
		if code, ok := sshE.GetExitCode(); ok {
			os.Exit(code)
		}

		os.Exit(1)
	}

	if e, ok := m.cmdErr.(exiterror.Error); ok {
		os.Exit(e.Code())
	}

	os.Exit(1)
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
		"error_type": errorhandler.GetTypes(m.cmdErr),
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

	if commandPath == "lcp metrics usage reset" {
		// Skip storing "lcp metrics usage reset" on the analytics log
		// otherwise this would recreate the file just after removal
		return
	}

	metrics.Rec(m.config, metrics.Event{
		Type:  "cmd",
		Text:  commandPath,
		Tags:  m.getCommandFlags(),
		Extra: m.getCommandErrorDetails(),
	})
}

func (m *mainProgram) maybeSubmitAnalyticsReport() {
	if !isCommand("metrics") && !isCommand("uninstall") {
		if err := metrics.SubmitEventuallyOnBackground(m.config); err != nil {
			_, _ = fmt.Fprintf(os.Stderr,
				"Error trying to submit analytics on background: %v\n",
				errorhandler.Handle(err))
		}
	}
}

type configLoader struct {
	reload bool
	wectx  config.Context
}

func (cl *configLoader) loadConfig() {
	var path = filepath.Join(userhome.GetHomeDir(), ".lcp")

	wectx, err := config.Setup(path)
	we.WithContext(&wectx)
	cl.wectx = wectx

	if err != nil {
		printError(errorhandler.Handle(err))
		verbose.PrintDeferred()
		os.Exit(1)
	}

	var conf = wectx.Config()
	var params = conf.GetParams()

	if params.NoColor {
		color.NoColor = true
		conf.SetParams(params)
	}
}

func (cl *configLoader) checkPastVersion() {
	var wectx = cl.wectx
	var conf = wectx.Config()
	var params = conf.GetParams()

	if params.PastVersion != "" {
		update.ApplyTransitions(params.PastVersion)
		params.PastVersion = ""
		conf.SetParams(params)
		cl.reload = true
	}
}

func (cl *configLoader) checkAnalytics() {
	var wectx = cl.wectx
	var conf = wectx.Config()
	var params = conf.GetParams()

	if !params.EnableAnalytics || params.AnalyticsID != "" {
		return
	}

	if err := metrics.Enable(cl.wectx.Config()); err != nil {
		verbose.Debug("Error trying to fix enabled metrics without analytics ID: " + err.Error())
	}
}

func (cl *configLoader) applyChanges() {
	if !cl.reload {
		return
	}

	if err := cl.wectx.Config().Save(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", errorhandler.Handle(err))
		verbose.PrintDeferred()
		os.Exit(1)
	}

	cl.loadConfig()
}

func (cl *configLoader) load() *config.Config {
	cl.loadConfig()
	cl.checkAnalytics()
	cl.checkPastVersion()
	cl.applyChanges()
	return cl.wectx.Config()
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
	errorhandler.CommandName = strings.Join(os.Args[1:], " ")
}

func panickingListener(panicking *bool) {
	if !*panicking {
		return
	}

	errorhandler.Info()
	// don't recover from panic to get more context
	// to avoid having to handle it
	// unless we find out it is really useful later
	var wectx = we.Context()
	metrics.Rec(wectx.Config(), metrics.Event{
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
		_, _ = fmt.Fprintln(os.Stderr, fancy.Error(`Run "lcp --help" for usage.`))
	}
}
