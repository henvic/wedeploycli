package servertest

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/wedeploy/api-go"
)

var (
	// Mux is the mock server HTTP Request multiplexer
	Mux *http.ServeMux

	// IntegrationMux is the mock server HTTP Request multiplexer for integration tests
	IntegrationMux *http.ServeMux

	// IntegrationServer is the mock server for integration tests
	IntegrationServer *httptest.Server

	defaultHTTPClient *http.Client
	server            *httptest.Server
)

// Setup the mock server and setup WeDeploy client setup with it
func Setup() {
	Mux = http.NewServeMux()
	server = httptest.NewServer(Mux)

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	defaultHTTPClient = wedeploy.Client
	wedeploy.Client = &http.Client{Transport: transport}
}

// SetupIntegration sets up the integration tests mock server
func SetupIntegration() {
	IntegrationMux = http.NewServeMux()
	IntegrationServer = httptest.NewServer(IntegrationMux)
}

// Teardown the mock server and teardown WeDeploy client
func Teardown() {
	wedeploy.Client = defaultHTTPClient
	server.Close()
	Mux = nil
	server = nil
}

// TeardownIntegration teardowns the integration tests mock server
func TeardownIntegration() {
	IntegrationServer.Close()
	IntegrationMux = nil
	IntegrationServer = nil
}
