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
	"strings"
	"time"

	"github.com/henvic/browser"
	"github.com/wedeploy/cli/cmd/canceled"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/fancy"
	"github.com/wedeploy/cli/formatter"
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
	var (
		username string
		password string
		token    string
		err      error

		remoteAddress = a.wectx.InfrastructureDomain()
	)

	fmt.Println(fancy.Info("Alert     You need a WeDeploy password for authenticating without opening your browser." +
		"\n          If you created your WeDeploy account by connecting to your Google or GitHub account," +
		"\n          make sure you set up a password to continue."))
	fmt.Println(color.Format(color.FgHiYellow, "\n            Open this URL in your browser for creating a password:"))
	fmt.Println(color.Format(color.FgHiBlack, fmt.Sprintf("            %v%v/password/reset\n", defaults.DashboardURLPrefix, remoteAddress)))

	fmt.Println(fancy.Question("Type your credentials for logging in. Your email: ") + color.Format(color.FgHiMagenta, "[ex: user@domain.com]"))
promptForUsername:
	if username, err = fancy.Prompt(); err != nil {
		return err
	}

	if validEmailAddress, invalidEmailAddressMsg := validateEmail(username); !validEmailAddress {
		fmt.Fprintf(os.Stderr, "%s\n", fancy.Error(invalidEmailAddressMsg))
		goto promptForUsername
	}

	fmt.Print(fancy.Question("Great! Now, your password:\n"))
promptForPassword:
	if password, err = fancy.HiddenPrompt(); err != nil {
		return err
	}

	fmt.Println(color.Format(color.FgHiBlack, "●●●●●●●●●●"))
	if validPassword, invalidPasswordMsg := validatePassword(password); !validPassword {
		fmt.Fprintf(os.Stderr, "%s\n", fancy.Error(invalidPasswordMsg))
		goto promptForPassword
	}

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

	token, err = ba.GetOAuthToken(ctx)
	a.maybePrintReceivedToken(token)

	if err != nil {
		a.msg.StopText(fancy.Error("Authentication failed [1/2]"))
		return err
	}

	return a.saveUser(username, token)
}

func (a *Authentication) tryStdinToken() (bool, error) {
	file := os.Stdin
	fi, err := file.Stat()

	// Different systems treat Stdin differently
	// On Ubuntu (Linux), the stdin size is zero even if all
	// content was already piped, say with:
	// echo foo | we login
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

	token, err := usertoken.ParseUnsignedJSONWebToken(maybe)

	if err != nil {
		return true, err
	}

	a.wlm = waitlivemsg.New(nil)
	a.msg = waitlivemsg.NewMessage("Authenticating [1/2]")
	a.wlm.AddMessage(a.msg)
	go a.wlm.Wait()
	defer a.wlm.Stop()
	return true, a.saveUser(token.Email, maybe)
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

	fmt.Println("WeDeploy requires your browser for authenticating.")
	yes, err := fancy.Boolean("Open your browser and authenticate?")

	if err != nil {
		return err
	}

	if !yes {
		return canceled.CancelCommand("Login canceled.")
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
	var remote = conf.Remotes[a.wectx.Remote()]

	var buf = &bytes.Buffer{}
	fmt.Fprintln(buf, fancy.Success(fmt.Sprintf("Authentication completed in %s [2/2]", timehelper.RoundDuration(duration, time.Second))))
	fmt.Fprintln(buf, fancy.Success(fmt.Sprintf(`You're logged in as "`+
		color.Format(color.Reset, color.Bold, username)+
		color.Format(color.FgHiGreen, `" on "`)+
		color.Format(color.Reset, color.Bold, remote.Infrastructure)+
		color.Format(color.FgHiGreen, `".`))))

	if a.TipCommands {
		a.printTipCommands(buf)
	}
	a.msg.StopText(buf.String())
}

func (a *Authentication) printTipCommands(buf *bytes.Buffer) {
	fmt.Fprintln(buf, fancy.Info("Check out some useful commands in case you wanna start learning the CLI:\n"))
	tw := formatter.NewTabWriter(buf)
	fmt.Fprintln(tw, color.Format(color.FgHiBlack, "  Command\t  Description"))
	fmt.Fprintln(tw, "  we\t  Show list of all commands available in WeDeploy CLI")
	fmt.Fprintln(tw, "  we deploy\t  Deploy your services")
	fmt.Fprintln(tw, "  we docs\t  Open docs on your browser")
	_ = tw.Flush()
	fmt.Fprint(buf, fancy.Info("\nType a command and press Enter to execute it."))
}

func (a *Authentication) saveUser(username, token string) (err error) {
	var conf = a.wectx.Config()
	var remote = conf.Remotes[a.wectx.Remote()]
	remote.Username = username
	remote.Token = token
	remote.Infrastructure = a.Domains.Infrastructure
	remote.Service = a.Domains.Service

	conf.Remotes[a.wectx.Remote()] = remote

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
