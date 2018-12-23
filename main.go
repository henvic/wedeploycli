/*
cli.cmd

	https://github.com/wedeploy/cli

*/

package main

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/wedeploy/cli/cmd"
	"github.com/wedeploy/cli/cmd/gitcredentialhelper"
	"github.com/wedeploy/cli/envs"
	wedeploy "github.com/wedeploy/wedeploy-sdk-go"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/client"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

func maybeSetCustomTimezone() {
	timezone := os.Getenv(envs.TZ)

	if timezone == "" {
		return
	}

	l, err := time.LoadLocation(timezone)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failure setting a custom timezone: %+v\n", err)
		return
	}

	time.Local = l
}

func maybeShortcutCredentialHelper() {
	if len(os.Args) < 2 || os.Args[1] != "git-credential-helper" {
		return
	}

	var err = gitcredentialhelper.Run(os.Args)

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func maybeSetSkipTLSVerification() {
	skipTLSVerification := os.Getenv(envs.SkipTLSVerification)

	if skipTLSVerification == "" {
		return
	}

	wedeployClient := wedeploy.Client()
	dt := http.DefaultTransport.(*http.Transport)

	// create new Transport that ignores self-signed SSL
	t := &http.Transport{
		// deep copy values from net/http DefaultTransport
		Proxy:                 dt.Proxy,
		DialContext:           dt.DialContext,
		MaxIdleConns:          dt.MaxIdleConns,
		IdleConnTimeout:       dt.IdleConnTimeout,
		ExpectContinueTimeout: dt.ExpectContinueTimeout,
		TLSHandshakeTimeout:   dt.TLSHandshakeTimeout,

		// With an unsafe TLS config
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	c := &http.Client{
		Transport: t,
	}

	// override only the wedeploy HTTP client, instead of http.DefaultTransport,
	// as it is less risky than for any clients
	wedeployClient.SetHTTP(c)

	// Install it as default client for https URLs.
	client.InstallProtocol("https", githttp.NewClient(c))
}

func main() {
	maybeSetCustomTimezone()
	maybeSetSkipTLSVerification()

	rand.Seed(time.Now().UTC().UnixNano())

	maybeShortcutCredentialHelper()
	cmd.Execute()
}
