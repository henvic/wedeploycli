package instances

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

var (
	wectx  config.Context
	client *Client
)

func TestMain(m *testing.M) {
	var err error
	wectx, err = config.Setup("mocks/.liferaycli")

	if err != nil {
		panic(err)
	}

	if err := wectx.SetEndpoint(defaults.CloudRemote); err != nil {
		panic(err)
	}

	client = New(wectx)
	os.Exit(m.Run())
}

func TestList(t *testing.T) {
	servertest.Setup()

	var want = []Instance{
		Instance{
			InstanceID: "abc00000",
			ServiceID:  "mail",
			ProjectID:  "project",
		},
		Instance{
			InstanceID: "abcd1234",
			ServiceID:  "mail",
			ProjectID:  "project",
		},
	}

	servertest.Mux.HandleFunc("/instances",
		func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()

			if q.Get("instanceId") != "abc" {
				t.Error("instanceId mismatch")
			}

			if q.Get("projectId") != "project" {
				t.Error("projectId mismatch")
			}

			if q.Get("serviceId") != "mail" {
				t.Error("serviceId mismatch")
			}

			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			_, _ = fmt.Fprintf(w, "%v", tdata.FromFile("mocks/instances_response.json"))
		})

	f := Filter{
		InstanceID: "abc",
		ServiceID:  "mail",
		ProjectID:  "project",
	}

	var got, err = client.List(context.Background(), f)

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	// if !reflect.DeepEqual(want, got) {
	if fmt.Sprintf("%+v", got) != fmt.Sprintf("%+v", want) {
		t.Errorf("List does not match with wanted structure.")
		t.Errorf(pretty.Compare(want, got))
	}

	servertest.Teardown()
}
