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

func maybeSetSkipTLSVerification() {
	skipTLSVerification := os.Getenv(envs.SkipTLSVerification)

	if skipTLSVerification == "" {
		return
	}

	wedeployClient := wedeploy.Client()
	dt := http.DefaultTransport.(*http.Transport)

	// create new Transport that ignores self-signed SSL
	httpClientWithSelfSignedTLS := &http.Transport{
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

	// override only the wedeploy HTTP client, instead of http.DefaultTransport,
	// as it is less risky than for any clients
	wedeployClient.SetHTTP(&http.Client{
		Transport: httpClientWithSelfSignedTLS,
	})
}

func main() {
	maybeSetCustomTimezone()
	maybeSetSkipTLSVerification()

	rand.Seed(time.Now().UTC().UnixNano())

	var args = os.Args

	if len(args) >= 2 && args[1] == "git-credential-helper" {
		var err = gitcredentialhelper.Run(args)

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		return
	}

	cmd.Execute()
}
