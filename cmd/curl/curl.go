package curl

// NOTE: curl's --url is not used due to conflicting we --url flag
// However, this code considers it (though the code wounever be executed),
// for the sake of completion [and if things change].

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/wedeploy/cli/cmd/curl/internal/curlargs"
	"github.com/wedeploy/cli/cmd/internal/we"
	"github.com/wedeploy/cli/cmdflagsfromhost"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/prettyjson"
	"github.com/wedeploy/cli/verbose"
)

// CurlCmd do curl requests using the user credential
var CurlCmd = &cobra.Command{
	Use:   "curl",
	Short: "Do requests with curl",
	Long: `Do requests with curl
Requests are piped to curl with credentials attached and paths expanded.
Pattern: we curl [curl options...] <url>
Use "curl --help" to see curl usage options.
`,
	Example: `  we curl /projects
  we curl /plans/user
  we curl https://api.wedeploy.com/projects`,
	// maybe --pretty=false to disable pipe, should add example
	RunE:               (&curlRunner{}).run,
	Hidden:             true,
	DisableFlagParsing: true,
}

// EnableCmd for enabling using "we curl"
var EnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable curl commands",
	RunE:  enableRun,
	Args:  cobra.NoArgs,
}

// DisableCmd for disabling using "we curl"
var DisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable curl commands",
	RunE:  disableRun,
	Args:  cobra.NoArgs,
}

var setupHost = cmdflagsfromhost.SetupHost{
	Pattern: cmdflagsfromhost.RemotePattern,
}

var print bool
var noPretty bool

func init() {
	CurlCmd.AddCommand(EnableCmd)
	CurlCmd.AddCommand(DisableCmd)

	CurlCmd.Flags().BoolVar(
		&print,
		"print",
		false,
		"Print command instead of invoking")
	CurlCmd.Flags().BoolVar(
		&noPretty,
		"no-pretty",
		false,
		"Don't pretty print JSON")
	setupHost.Init(CurlCmd)
}

type argsF struct {
	input []string
	pos   int

	weArgs   []string
	curlArgs []string

	pfs []*pflag.Flag
}

func (af *argsF) maybeGetBoolArgument() (is bool) {
	arg := af.input[af.pos]

	// -H might be either --long-help or curl header
	if arg == "-H" && (af.pos+1 >= len(af.input) ||
		strings.HasPrefix(af.input[af.pos+1], "-")) {
		af.weArgs = append(af.weArgs, arg)
		return true
	}

	for _, p := range af.pfs {
		if arg == "--"+p.Name || arg == "-"+p.Shorthand {
			// -H requires a special treatment, given above
			if arg == "-H" {
				continue
			}

			af.weArgs = append(af.weArgs, arg)

			if p.Value.Type() != "bool" && af.pos+1 < len(af.input) {
				af.weArgs = append(af.weArgs, af.input[af.pos+1])
				af.pos++
			}

			return true
		}
	}

	return false
}

func (cr *curlRunner) parseArguments() (weArgs, curlArgs []string) {
	// ignore "we curl"
	var commandLength = len(strings.Split(cr.cmd.CommandPath(), " "))

	if len(os.Args) <= commandLength {
		return []string{}, []string{}
	}

	var af = argsF{
		input: os.Args[commandLength:],
	}

	cr.cmd.Flags().VisitAll(func(f *pflag.Flag) {
		af.pfs = append(af.pfs, f)
	})

	for {
		if af.pos >= len(af.input) {
			break
		}

		if got := af.maybeGetBoolArgument(); !got {
			af.curlArgs = append(af.curlArgs, af.input[af.pos])
		}

		af.pos++
	}

	return af.weArgs, af.curlArgs
}

func isSafeInfrastructureURL(wectx config.Context, param string) bool {
	u, err := url.Parse(param)
	extractedHost := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	if err != nil || extractedHost == wectx.Infrastructure() {
		return true
	}

	return false
}

// UnsafeURLError is used when a URL is dangerous to add on the curl command
type UnsafeURLError struct {
	url string
}

func (u UnsafeURLError) Error() string {
	return fmt.Sprintf("refusing due to possibly unsafe URL value: %v", u.url)
}

func extractRemoteFromFullPath(wectx config.Context, fullPath string) (string, error) {
	var conf = wectx.Config()
	u, err := url.Parse(fullPath)

	if err != nil {
		return "", err
	}

	for key, r := range conf.Remotes {
		i, err := url.Parse(r.InfrastructureServer())

		if err != nil {
			continue
		}

		if u.Host == i.Host && u.Scheme == i.Scheme {
			return key, err
		}
	}

	return "", nil
}

func extractAlternativeRemote(wectx config.Context, params []string) (string, error) {
	var alternative string

	for i, p := range params {
		if strings.HasPrefix(p, "-") || strings.HasPrefix(p, "/") {
			continue
		}

		if i == 0 {
			var remote, err = extractRemoteFromFullPath(wectx, p)

			if err != nil {
				continue
			}

			if alternative != "" && alternative != remote {
				return "", UnsafeURLError{p}
			}

			alternative = remote
			continue
		}

		var iss = curlargs.IsCLIStringArgument(params[i-1])

		if !iss || (iss && params[i-1] == "--url") {
			var remote, err = extractRemoteFromFullPath(wectx, p)

			if err != nil {
				continue
			}

			if alternative != "" && alternative != remote {
				return "", UnsafeURLError{p}
			}

			alternative = remote
			continue
		}
	}

	// at the end, check if it's not the same and ignore if it is
	if alternative == wectx.Remote() {
		alternative = ""
	}

	return alternative, nil
}

func expandPathsToFullRequests(wectx config.Context, params []string) ([]string, error) {
	var out []string

	for i, p := range params {
		if strings.HasPrefix(p, "-") {
			out = append(out, p)
			continue
		}

		// let's try to expand URLs (e.g., "we curl /projects" should work)
		if strings.HasPrefix(p, "/") {
			if i == 0 {
				out = append(out, fmt.Sprintf("%v%v", wectx.Infrastructure(), p))
				continue
			}

			// expand string arguments, except for --url
			if curlargs.IsCLIStringArgument(params[i-1]) && params[i-1] != "--url" {
				out = append(out, p)
				continue
			}

			out = append(out, fmt.Sprintf("%v%v", wectx.Infrastructure(), p))
			continue
		}

		if i == 0 {
			if !isSafeInfrastructureURL(wectx, p) {
				return nil, UnsafeURLError{p}
			}

			out = append(out, p)
			continue
		}

		// expand string arguments, except for --url
		if curlargs.IsCLIStringArgument(params[i-1]) {
			if params[i-1] == "--url" && !isSafeInfrastructureURL(wectx, p) {
				return nil, UnsafeURLError{p}
			}

			out = append(out, p)
			continue
		}

		if !isSafeInfrastructureURL(wectx, p) {
			return nil, UnsafeURLError{p}
		}

		out = append(out, p)
	}

	return out, nil
}

func enableRun(cmd *cobra.Command, args []string) (err error) {
	var wectx = we.Context()
	var conf = wectx.Config()

	conf.EnableCURL = true
	return conf.Save()
}

func disableRun(cmd *cobra.Command, args []string) (err error) {
	var wectx = we.Context()
	var conf = wectx.Config()

	conf.EnableCURL = false
	return conf.Save()
}

func maybeChangeRemote(wectx config.Context, cmd *cobra.Command, weArgs []string, alternative string) error {
	// if no change on remote, shortcuit it
	if alternative == "" {
		return nil
	}

	if alternative != "" && (cmd.Flag("remote").Changed || cmd.Flag("url").Changed) {
		return fmt.Errorf(`ambiguous remote options: "%s" and "%s"`,
			setupHost.Remote(), alternative)
	}

	if err := cmd.Flag("remote").Value.Set(alternative); err != nil {
		return errwrap.Wrapf("can't override remote value: {{err}}", err)
	}

	return wectx.SetEndpoint(setupHost.Remote())
}

type curlRunner struct {
	cmd *cobra.Command
}

func (cr *curlRunner) run(cmd *cobra.Command, args []string) error {
	cr.cmd = cmd

	// Let's try to avoid verbose messages from the CLI
	// get in the way of verbose of the curl command
	defer func() {
		verbose.Enabled = false
	}()

	var wectx = we.Context()

	var weArgs, curlArgs = cr.parseArguments()

	cmd.DisableFlagParsing = false
	if err := cmd.ParseFlags(weArgs); err != nil {
		return err
	}

	if cmd.Flag("help").Value.String() == "true" ||
		cmd.Flag("long-help").Value.String() == "true" || len(args) == 0 {
		return cmd.Help()
	}

	if !wectx.Config().EnableCURL {
		_, _ = fmt.Fprintln(os.Stderr,
			`This command is not enabled by default as it might be dangerous for security.
Using it might make you inadvertently expose private data. Continue at your own risk.`)

		return fmt.Errorf(`You must enable this command first with "%v"`,
			EnableCmd.CommandPath())
	}

	if err := setupHost.Process(context.Background(), wectx); err != nil {
		return err
	}

	alternative, err := extractAlternativeRemote(wectx, curlArgs)

	if err != nil {
		return err
	}

	if err = maybeChangeRemote(wectx, cmd, weArgs, alternative); err != nil {
		return err
	}

	if verbose.Enabled {
		curlArgs = append(curlArgs, "--verbose")
	}

	curlArgs, err = expandPathsToFullRequests(wectx, curlArgs)

	if err != nil {
		return err
	}

	token := wectx.Token()

	if token != "" {
		curlArgs = append(curlArgs, "-H")
		curlArgs = append(curlArgs, fmt.Sprintf("Authorization: Bearer %s", token))
	}

	curlArgs = append([]string{"-sS"}, curlArgs...)

	if print {
		printCURLCommand(curlArgs)
		return nil
	}

	if !noPretty {
		return curlPretty(context.Background(), curlArgs)
	}

	return curl(context.Background(), curlArgs)
}

func curl(ctx context.Context, params []string) error {
	verbose.Debug(fmt.Sprintf("Running curl %v", strings.Join(params, " ")))

	var cmd = exec.CommandContext(ctx, "curl", params...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func curlPretty(ctx context.Context, params []string) error {
	verbose.Debug(fmt.Sprintf("Running curl %v", strings.Join(params, " ")))

	var (
		buf    bytes.Buffer
		bufErr bytes.Buffer
	)

	var cmd = exec.CommandContext(ctx, "curl", params...)
	cmd.Stdin = os.Stdin

	cmd.Stderr = io.MultiWriter(&bufErr, os.Stderr)
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return err
	}

	maybePrettyPrintJSON(bufErr.Bytes(), buf.Bytes())
	return nil
}

func maybePrettyPrintJSON(headersOrErr, body []byte) {
	if verbose.Enabled &&
		!bytes.Contains(bytes.ToLower(headersOrErr), []byte("\n< content-type: application/json")) {
		fmt.Println(string(body))
		return
	}

	fmt.Print(string(prettyjson.Pretty(body)))
}

func printCURLCommand(args []string) {
	fmt.Printf("curl")

	for _, a := range args {
		if strings.ContainsRune(a, ' ') {
			fmt.Printf(` "%s"`, a)
			break
		}

		fmt.Printf(" %s", a)
	}

	fmt.Printf("\n")
}
