package login

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/henvic/browser"
	"github.com/wedeploy/cli/cmd/canceled"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/figures"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/isterm"
	"github.com/wedeploy/cli/loginserver"
	"github.com/wedeploy/cli/status"
	"github.com/wedeploy/cli/timehelper"
	"github.com/wedeploy/cli/usertoken"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/waitlivemsg"
)

func validateEmail(email string) (bool, error) {
	if len(email) == 0 {
		return false, errors.New("please enter your email")
	}

	var index = strings.Index(email, "@")

	if index == -1 {
		return false, errors.New(`please enter your full email address, including the "@"`)
	}

	if index+1 == len(email) {
		return false, errors.New(`do not forget the part after the "@"`)
	}

	return true, nil
}

func validatePassword(password string) (bool, error) {
	if len(password) == 0 {
		return false, errors.New("please enter your password")
	}

	return true, nil
}

// Authentication service
type Authentication struct {
	NoLaunchBrowser bool
	Domains         status.Domains
	TipCommands     bool
	wectx           config.Context
	wlm             *waitlivemsg.WaitLiveMsg
	msg             *waitlivemsg.Message
}

func (a *Authentication) basicAuthLogin(ctx context.Context) error {
	var remoteAddress = a.wectx.InfrastructureDomain()

	fmt.Println(fancy.Info("Alert     You need a Liferay Cloud password for authenticating without opening your browser." +
		"\n          If you created your Liferay Cloud account by using OAuth," +
		"\n          make sure you set up a password to continue."))
	fmt.Println(color.Format(color.FgHiYellow, "\n            Open this URL in your browser for creating a password:"))
	fmt.Println(color.Format(color.FgHiBlack, fmt.Sprintf("            %v%v/password/reset\n", defaults.DashboardURLPrefix, remoteAddress)))

	fmt.Println(fancy.Question("Type your credentials for logging in. Your email: ") + color.Format(color.FgHiBlack, "[ex: user@domain.com]"))
promptForUsername:

	username, err := fancy.Prompt()

	if err != nil {
		return err
	}

	if validEmailAddress, invalidEmailAddressMsg := validateEmail(username); !validEmailAddress {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", fancy.Error(invalidEmailAddressMsg))
		goto promptForUsername
	}

	fmt.Print(fancy.Question("Great! Now, your password:\n"))
promptForPassword:
	password, err := fancy.HiddenPrompt()

	if err != nil {
		return err
	}

	fmt.Println(color.Format(color.FgHiBlack, "●●●●●●●●●●"))
	if validPassword, invalidPasswordMsg := validatePassword(password); !validPassword {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", fancy.Error(invalidPasswordMsg))
		goto promptForPassword
	}

	return a.loginWithCredentials(ctx, username, password)
}

func (a *Authentication) loginWithCredentials(ctx context.Context, username, password string) error {
	a.wlm = waitlivemsg.New(nil)
	a.msg = waitlivemsg.NewMessage("Authenticating [1/2]")
	a.wlm.AddMessage(a.msg)
	go a.wlm.Wait()
	defer a.wlm.Stop()

	var ba = loginserver.BasicAuth{
		Username: username,
		Password: password,
		Context:  a.wectx,
	}

	var token, err = ba.GetOAuthToken(ctx)
	a.maybePrintReceivedToken(token)

	if err != nil {
		a.msg.StopText(fancy.Error("Authentication failed [1/2]"))
		return err
	}

	return a.saveUser(username, token)
}

func (a *Authentication) tryStdinToken() (bool, error) {
	// trade-off: --no-tty is required for piping tokens on some Windows """shell subsystems"""
	// see issue https://github.com/wedeploy/cli/issues/435
	// error first appeared in commit 4d217d2324825714bf6fa35d988502692f0d7925
	if runtime.GOOS == "windows" &&
		(os.Getenv("OSTYPE") == "cygwin" ||
			strings.Contains(os.Getenv("MSYSTEM_CHOST"), "mingw") ||
			strings.Contains(os.Getenv("MINGW_CHOST"), "mingw")) {
		verbose.Debug("INFO: --no-tty is required to pipe credentials values such as tokens using STDIN on some Windows environments")

		if !isterm.NoTTY {
			return false, nil
		}
	}

	file := os.Stdin
	fi, err := file.Stat()

	// Different systems treat Stdin differently
	// On Ubuntu (Linux), the stdin size is zero even if all
	// content was already piped, say with:
	// echo foo | lcp login
	// On Darwin (macOS), this is not the case.
	// See http://learngowith.me/a-better-way-of-handling-stdin/

	if fi.Size() != 0 {
		goto skipToStdin
	}

	if err != nil || fi.Mode()&os.ModeCharDevice != 0 {
		return false, nil
	}

skipToStdin:
	reader := bufio.NewReader(file)
	maybe, err := reader.ReadString('\n')

	if err != nil && err != io.EOF {
		return false, nil
	}

	maybe = strings.TrimSuffix(maybe, "\n")

	if sep := strings.Index(maybe, " "); sep != -1 {
		return true, a.loginWithCredentials(context.Background(), maybe[:sep], maybe[sep+1:])
	}

	return true, a.loginWithToken(maybe)
}

func (a *Authentication) loginWithToken(token string) error {
	wt, err := usertoken.ParseUnsignedJSONWebToken(token)

	if err != nil {
		return err
	}

	a.wlm = waitlivemsg.New(nil)
	a.msg = waitlivemsg.NewMessage("Authenticating [1/2]")
	a.wlm.AddMessage(a.msg)
	go a.wlm.Wait()
	defer a.wlm.Stop()
	return a.saveUser(wt.Email, token)
}

// Run authentication process
func (a *Authentication) Run(ctx context.Context, c config.Context) error {
	a.wectx = c
	statusClient := status.New(c)

	s, err := statusClient.UnsafeGet(ctx)

	if err != nil {
		return err
	}

	a.Domains = s.Domains

	if stdin, stdinErr := a.tryStdinToken(); stdin {
		return stdinErr
	}

	if a.NoLaunchBrowser {
		return a.basicAuthLogin(ctx)
	}

	yes, err := fancy.Boolean("Open your browser and authenticate?")

	if err != nil {
		return err
	}

	if !yes {
		return canceled.CancelCommand("login canceled")
	}

	return a.browserWorkflowAuth()
}

func (a *Authentication) maybeOpenBrowser(loginURL string) {
	if verbose.Enabled {
		a.wlm.AddMessage(waitlivemsg.NewMessage("Login URL: " + loginURL))
	}

	time.Sleep(710 * time.Millisecond)

	if err := browser.OpenURL(loginURL); err != nil {
		errMsg := &waitlivemsg.Message{}
		errMsg.StopText(fmt.Sprintf("%v", err))
		a.wlm.AddMessage(errMsg)

		if !verbose.Enabled {
			a.wlm.AddMessage(waitlivemsg.NewMessage("Open URL: (can't open automatically) " + loginURL))
		}
	}
}

func (a *Authentication) browserWorkflowAuth() error {
	a.wlm = waitlivemsg.New(nil)
	a.msg = waitlivemsg.NewMessage("Waiting for authentication via browser [1/2]\n" +
		fancy.Tip("^C to cancel"))
	a.wlm.AddMessage(a.msg)
	go a.wlm.Wait()
	defer a.wlm.Stop()
	var service = &loginserver.Service{
		Infrastructure: a.Domains.Infrastructure,
	}
	var host, err = service.Listen(context.Background())

	if err != nil {
		a.msg.StopText(fancy.Error("Authentication failed [1/2]"))
		return err
	}

	var loginURL = fmt.Sprintf("%s%s%s%s",
		defaults.DashboardURLPrefix,
		a.wectx.InfrastructureDomain(),
		"/login?redirect_uri=",
		url.QueryEscape(host))

	a.maybeOpenBrowser(loginURL)

	if err = service.Serve(); err != nil {
		a.msg.StopText(fancy.Error("Authentication failed [1/2]"))
		return err
	}

	var username, token, tokenErr = service.Credentials()
	a.maybePrintReceivedToken(token)

	if tokenErr != nil {
		a.msg.StopText(fancy.Error("Authentication failed [1/2]"))
		return tokenErr
	}

	return a.saveUser(username, token)
}

func (a *Authentication) success(username string) {
	var duration = a.wlm.Duration()
	var conf = a.wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes
	var remote = rl.Get(a.wectx.Remote())

	var buf = &bytes.Buffer{}
	_, _ = fmt.Fprintf(buf, "%s Authentication completed in %s [2/2]\n", figures.Tick, timehelper.RoundDuration(duration, time.Second))
	_, _ = fmt.Fprintf(buf, "You're logged in as \"%s\" on \"%s\".\n",
		color.Format(color.Reset, color.Bold, username),
		color.Format(color.Reset, color.Bold, remote.Infrastructure))

	if a.TipCommands {
		a.printTipCommands(buf)
	}
	a.msg.StopText(buf.String())
}

func (a *Authentication) printTipCommands(buf *bytes.Buffer) {
	_, _ = fmt.Fprintln(buf, fancy.Info("See some of the useful commands you can start using on the Liferay Cloud Platform CLI.\n"))
	tw := formatter.NewTabWriter(buf)
	_, _ = fmt.Fprintln(tw, color.Format(color.FgHiBlack, "  Command\t     Description"))
	_, _ = fmt.Fprintln(tw, "  lcp\tShow list of all commands available in Liferay Cloud Platform CLI")
	_, _ = fmt.Fprintln(tw, "  lcp deploy\tDeploy your services")
	_, _ = fmt.Fprintln(tw, "  lcp docs\tOpen docs on your browser")
	_ = tw.Flush()
	_, _ = fmt.Fprint(buf, fancy.Info("\nType a command and press Enter to execute it."))
}

func (a *Authentication) saveUser(username, token string) (err error) {
	var conf = a.wectx.Config()
	var params = conf.GetParams()
	var rl = params.Remotes
	var remote = rl.Get(a.wectx.Remote())
	remote.Username = username
	remote.Token = token
	remote.Infrastructure = a.Domains.Infrastructure
	remote.Service = a.Domains.Service

	rl.Set(a.wectx.Remote(), remote)

	if err = a.wectx.SetEndpoint(a.wectx.Remote()); err != nil {
		a.msg.StopText(fancy.Error("Authentication failed [1/2]"))
		return err
	}

	if err = conf.Save(); err != nil {
		a.msg.StopText(fancy.Error("Authentication failed [1/2]"))
		return err
	}

	a.success(username)
	return nil
}

func (a *Authentication) maybePrintReceivedToken(token string) {
	if verbose.Enabled {
		tokenMsg := &waitlivemsg.Message{}
		tokenMsg.StopText("Token: " + verbose.SafeEscape(token))
		a.wlm.AddMessage(tokenMsg)
	}
}
