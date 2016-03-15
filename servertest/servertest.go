package servertest

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/launchpad-project/api.go"
)

var (
	// Mux is the mock server HTTP Request multiplexer
	Mux *http.ServeMux

	defaultHTTPClient *http.Client
	server            *httptest.Server
)

// Setup the mock server and setup Launchpad client setup with it
func Setup() {
	Mux = http.NewServeMux()
	server = httptest.NewServer(Mux)

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	defaultHTTPClient = launchpad.Client
	launchpad.Client = &http.Client{Transport: transport}
}

// Teardown the mock server and teardown Launchpad client
func Teardown() {
	launchpad.Client = defaultHTTPClient
	server.Close()
}
