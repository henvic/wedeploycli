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

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/usertoken"
	"github.com/wedeploy/cli/verbose"
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

const redirectPage = `<html>
<body>
<form action="/authenticate" method="post" id="authenticate">
<input type="hidden" id="access_token" name="access_token" />
</form>
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
	s.ctx, s.ctxCancel = context.WithCancel(ctx)
	s.netListener, err = net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		return "", errwrap.Wrapf("Can not start authentication service: {{err}}", err)
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
		s.err = errwrap.Wrapf("Can not shutdown login service properly: {{err}}", err)
	}
	w.Done()
}

// Serve HTTP requests
func (s *Service) Serve() error {
	if s.netListener == nil {
		return errors.New("Server is not open yet")
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
	fmt.Fprintf(w, safeErrorPageTemplate, body)
}

func (s *Service) homeHandler(w http.ResponseWriter, r *http.Request) {
	referer, _ := url.Parse(r.Header.Get("Referer"))

	// this is a compromise
	var dashboard = defaults.DashboardAddressPrefix + s.Infrastructure
	if referer.Host != "" && referer.Host != dashboard {
		s.err = errors.New("Token origin is not from given dashboard")
		safeErrorHandler(w, "403 Forbidden", http.StatusForbidden)
		s.ctxCancel()
		return
	}

	fmt.Fprintln(w, redirectPage)
}

// ErrSignUpEmailConfirmation tells that sign up was canceled because user is signing up
var ErrSignUpEmailConfirmation = errors.New(`sign up on WeDeploy requested: try "we login" once you confirm your email`)

const signupRequestPseudoToken = "signup_requested"

func (s *Service) authenticateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.Header.Get("Referer") != s.serverAddress+"/" {
		s.err = errors.New("Authentication should have been POSTed and from a localhost origin")
		safeErrorHandler(w, "403 Forbidden", http.StatusForbidden)
		s.ctxCancel()
		return
	}

	var pferr = r.ParseForm()

	if pferr != nil {
		s.err = errwrap.Wrapf("Can not parse authentication form: {{err}}", pferr)
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

// OAuthTokenFromBasicAuth gets a token from a Basic Auth flow
func OAuthTokenFromBasicAuth(c config.Context, remoteAddress, username, password string) (token string, err error) {
	var apiClient = apihelper.New(c)

	var request = apiClient.URL(context.Background(), "/login")

	request.Form("email", username)
	request.Form("password", password)

	if err := apihelper.Validate(request, request.Post()); err != nil {
		return "", err
	}

	var data accessToken
	err = apihelper.DecodeJSON(request, &data)
	return data.AccessToken, err
}
