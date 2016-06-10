package servertest

import (
	"net/http"
	"testing"

	"github.com/wedeploy/api-go"
)

func TestSetupAndTeardown(t *testing.T) {
	var completed = false
	var wantStatusCode = 201
	var wedeployHTTPClient = wedeploy.Client

	Setup()

	if wedeployHTTPClient == wedeploy.Client {
		t.Error("Expected different WeDeploy HTTP Client instance")
	}

	Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		completed = true
	})

	req := wedeploy.URL("http://example.com/foo")

	if err := req.Post(); err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if !completed {
		t.Error("Request not completed")
	}

	if req.Response.StatusCode != wantStatusCode {
		t.Errorf("Wanted status code %v, got %v instead", wantStatusCode, req.Response.StatusCode)
	}

	Teardown()

	if wedeployHTTPClient != wedeploy.Client {
		t.Error("Expected same WeDeploy HTTP Client instance")
	}

	if server != nil {
		t.Error("Expected server reference to be gone")
	}

	if Mux != nil {
		t.Error("Expected mux reference to be gone")
	}
}

func TestSetupAndTeardownIntegration(t *testing.T) {
	var completed = false
	var wantStatusCode = 201

	SetupIntegration()

	IntegrationMux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		completed = true
	})

	req := wedeploy.URL(IntegrationServer.URL, "/foo")

	if err := req.Post(); err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	if !completed {
		t.Error("Request not completed")
	}

	if req.Response.StatusCode != wantStatusCode {
		t.Errorf("Wanted status code %v, got %v instead", wantStatusCode, req.Response.StatusCode)
	}

	TeardownIntegration()

	if IntegrationServer != nil {
		t.Error("Expected IntegrationServer reference to be gone")
	}

	if IntegrationMux != nil {
		t.Error("Expected IntegrationMux reference to be gone")
	}
}
