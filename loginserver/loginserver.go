package loginserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/henvic/wedeploycli/apihelper"
	"github.com/henvic/wedeploycli/config"
	"github.com/henvic/wedeploycli/defaults"
	"github.com/henvic/wedeploycli/usertoken"
	"github.com/henvic/wedeploycli/verbose"
)

// Service server for receiving JSON Web Token
type Service struct {
	Infrastructure string
	ctx            context.Context
	ctxCancel      context.CancelFunc
	netListener    net.Listener
	httpServer     *http.Server
	serverAddress  string
	temporaryToken string
	jwt            usertoken.JSONWebToken
	err            error
}

// BUG(henvic): Ajax could be used to avoid a small risk of the user being stuck in a white error
// when the authentication fails (at some implementation cost).
const redirectPage = `<html>
<body>
<form action="/authenticate" method="post" id="authenticate">
<input type="hidden" id="access_token" name="access_token" />
</form>
<noscript>
You need JavaScript enabled to complete the authentication. Enable it and try again.
</noscript>
<script>
var accessToken = document.location.hash.replace('#access_token=', '');
var rm = "#access_token=";

if (accessToken.indexOf(rm) === 0) {
	accessToken = accessToken.substr(rm.length)
}

document.querySelector("#access_token").value = accessToken;
document.querySelector("#authenticate").submit();
</script>
</body>
</html>`

// Listen for requests
func (s *Service) Listen(ctx context.Context) (address string, err error) {
	s.ctx, s.ctxCancel = context.WithTimeout(ctx, 15*time.Minute)
	s.netListener, err = net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		return "", errwrap.Wrapf("can't start authentication service: {{err}}", err)
	}

	s.serverAddress = fmt.Sprintf("http://localhost:%v",
		strings.TrimPrefix(
			s.netListener.Addr().String(),
			"127.0.0.1:"))

	return s.serverAddress, nil
}

func (s *Service) waitServer(w *sync.WaitGroup) {
	<-s.ctx.Done()
	var err = s.httpServer.Shutdown(s.ctx)
	if err != nil && err != context.Canceled {
		s.err = errwrap.Wrapf("can't shutdown login service properly: {{err}}", err)
	}
	w.Done()
}

// Serve HTTP requests
func (s *Service) Serve() error {
	if s.netListener == nil {
		return errors.New("server is not open yet")
	}

	s.httpServer = &http.Server{
		Handler: &handler{
			handler: s.httpHandler,
		},
	}

	var w sync.WaitGroup
	w.Add(1)
	go s.waitServer(&w)

	var serverErr = s.httpServer.Serve(s.netListener)

	if serverErr != http.ErrServerClosed {
		verbose.Debug(fmt.Sprintf("Error closing authentication server: %v", serverErr))
	}

	w.Wait()
	return s.err
}

func (s *Service) redirectToDashboard(w http.ResponseWriter, r *http.Request) {
	var page string

	switch s.err {
	case nil:
		page = "static/cli/login-success/"
	case ErrSignUpEmailConfirmation:
		page = "static/cli/login-requires-email-confirmation/"
	default:
		page = "static/cli/login-failure/"
	}

	var redirect = fmt.Sprintf("https://%v%v/%v", defaults.DashboardAddressPrefix, s.Infrastructure, page)
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

func (s *Service) httpHandler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		s.homeHandler(w, r)
	case "/authenticate":
		s.authenticateHandler(w, r)
	default:
		http.NotFound(w, r)
	}
}

const safeErrorPageTemplate = `<html>
<body>
<script>
document.location.hash = ""
</script>
%s
</body>
</html>
`

// safeErrorHandler basically clears any access_token from the fragment
// and does what http.Error does
func safeErrorHandler(w http.ResponseWriter, body string, code int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	w.WriteHeader(code)
	_, _ = fmt.Fprintf(w, safeErrorPageTemplate, body)
}

func (s *Service) homeHandler(w http.ResponseWriter, r *http.Request) {
	referer, _ := url.Parse(r.Header.Get("Referer"))

	// this is a compromise
	var dashboard = defaults.DashboardAddressPrefix + s.Infrastructure
	if referer.Host != "" && referer.Host != dashboard {
		s.err = errors.New("token origin is not from given dashboard")
		safeErrorHandler(w, "403 Forbidden", http.StatusForbidden)
		s.ctxCancel()
		return
	}

	_, _ = fmt.Fprintln(w, redirectPage)
}

// ErrSignUpEmailConfirmation tells that sign up was canceled because user is signing up
var ErrSignUpEmailConfirmation = errors.New(`sign up on Liferay Cloud requested: try "lcp login" once you confirm your email`)

const signupRequestPseudoToken = "signup_requested"

func (s *Service) authenticateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Header.Get("Referer") != s.serverAddress+"/" {
		s.err = errors.New("authentication should have been POSTed and from a localhost origin")
		safeErrorHandler(w, "403 Forbidden", http.StatusForbidden)
		s.ctxCancel()
		return
	}

	var pferr = r.ParseForm()

	if pferr != nil {
		s.err = errwrap.Wrapf("can't parse authentication form: {{err}}", pferr)
		safeErrorHandler(w, "400 Bad Request", http.StatusBadRequest)
		s.ctxCancel()
		return
	}

	s.temporaryToken = r.Form.Get("access_token")
	verbose.Debug("Access Token: " + verbose.SafeEscape(s.temporaryToken))

	switch s.temporaryToken {
	case signupRequestPseudoToken:
		s.err = ErrSignUpEmailConfirmation
	default:
		s.jwt, s.err = usertoken.ParseUnsignedJSONWebToken(s.temporaryToken)
	}

	s.redirectToDashboard(w, r)
	s.ctxCancel()
}

// Credentials for authenticated user or error, it blocks until the information is available
func (s *Service) Credentials() (username string, token string, err error) {
	<-s.ctx.Done()
	return s.jwt.Email, s.temporaryToken, s.err
}

type handler struct {
	handler func(w http.ResponseWriter, r *http.Request)
}

func (s *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler(w, r)
}

type accessToken struct {
	AccessToken string `json:"token"`
}

// BasicAuth credentials
type BasicAuth struct {
	Username string
	Password string
	Context  config.Context
}

// GetOAuthToken from a Basic Auth flow
func (b *BasicAuth) GetOAuthToken(ctx context.Context) (string, error) {
	var apiClient = apihelper.New(b.Context)
	var request = apiClient.URL(ctx, "/login")

	request.Form("email", b.Username)
	request.Form("password", b.Password)

	if err := apihelper.Validate(request, request.Post()); err != nil {
		return "", err
	}

	var data accessToken
	var err = apihelper.DecodeJSON(request, &data)
	return data.AccessToken, err
}
