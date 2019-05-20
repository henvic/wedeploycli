package servertest

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/wedeploy-sdk-go"
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

var mockServerURL *url.URL

type proxy struct{}

func (p *proxy) RoundTrip(r *http.Request) (w *http.Response, err error) {
	r.URL.Scheme = mockServerURL.Scheme
	r.URL.Host = mockServerURL.Host

	return (&http.Client{}).Do(r)
}

// Setup the mock server and setup Liferay Cloud client setup with it
func Setup() {
	Mux = http.NewServeMux()
	server = httptest.NewServer(Mux)
	var err error
	mockServerURL, err = url.Parse(server.URL)

	if err != nil {
		panic(errwrap.Wrapf("can't route to mock server: {{err}}", err))
	}

	client := wedeploy.Client()
	defaultHTTPClient = client.HTTP()
	client.SetHTTP(&http.Client{Transport: &proxy{}})
}

// SetupIntegration sets up the integration tests mock server
func SetupIntegration() {
	IntegrationMux = http.NewServeMux()
	IntegrationServer = httptest.NewServer(IntegrationMux)
}

// Teardown the mock server and teardown Liferay Cloud client
func Teardown() {
	wedeploy.Client().SetHTTP(defaultHTTPClient)
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
