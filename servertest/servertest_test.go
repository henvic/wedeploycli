package servertest

import (
	"net/http"
	"testing"

	"github.com/launchpad-project/api.go"
)

func TestSetupAndTeardown(t *testing.T) {
	var completed = false
	var wantStatusCode = 201
	var launchpadHTTPClient = launchpad.Client

	Setup()

	if launchpadHTTPClient == launchpad.Client {
		t.Error("Expected different Launchpad HTTP Client instance")
	}

	Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		completed = true
	})

	req := launchpad.URL("http://example.com/foo")
	req.Post()

	if !completed {
		t.Error("Request not completed")
	}

	if req.Response.StatusCode != wantStatusCode {
		t.Error("Wanted status code %v, got %v instead", wantStatusCode, req.Response.StatusCode)
	}

	Teardown()

	if launchpadHTTPClient != launchpad.Client {
		t.Error("Expected same Launchpad HTTP Client instance")
	}
}
