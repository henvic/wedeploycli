package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestList(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/projects",
		tdata.ServerJSONFileHandler("./mocks/list/projects_response.json"))

	servertest.IntegrationMux.HandleFunc(
		"/projects/wechat/services",
		tdata.ServerJSONFileHandler("./mocks/list/services_response.json"))

	var cmd = &Command{
		Args: []string{"list", "--remote", "local", "--no-color"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/list/want"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestListIncompatibleUseServiceRequiresProject(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"list", "--service", "service", "--remote", "local", "--no-color"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	cmd.Run()

	if update {
		tdata.ToFile("mocks/list/incompatible-use", cmd.Stderr.String())
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/list/incompatible-use"),
		ExitCode: 1,
	}

	e.Assert(t, cmd)
}

func TestListIncompatibleUseServiceRequiresProjectWithShortFlags(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"list", "-s", "service", "-r", "local", "--no-color"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	cmd.Run()

	if update {
		tdata.ToFile("mocks/list/incompatible-use", cmd.Stderr.String())
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/list/incompatible-use"),
		ExitCode: 1,
	}

	e.Assert(t, cmd)
}

func TestListServiceFromInsideProject(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc(
		"/projects/app",
		tdata.ServerJSONFileHandler("./mocks/home/bucket/project/service/projects_list"))

	servertest.IntegrationMux.HandleFunc(
		"/projects/app/services/service",
		tdata.ServerJSONFileHandler("./mocks/home/bucket/project/service/wedeploy.json"))

	servertest.IntegrationMux.HandleFunc(
		"/projects/app/services",
		tdata.ServerJSONFileHandler("./mocks/home/bucket/project/service/service_list.json"))

	var cmd = &Command{
		Args: []string{"list", "--service", "service", "--remote", "local", "--no-color"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/bucket/project",
	}

	var e = &Expect{
		Stdout:   tdata.FromFile("mocks/home/bucket/project/service/service_list_want"),
		ExitCode: 0,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestListServiceFromInsideProjectNotExists(t *testing.T) {
	defer Teardown()
	Setup()

	servertest.IntegrationMux.HandleFunc("/projects/app/services/service", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		w.WriteHeader(404)
		fmt.Fprintf(w, `{
    "status": 404,
    "message": "Not Found",
    "errors": [
        {
            "reason": "notFound",
            "message": "The requested operation failed because a resource associated with the request could not be found."
        }
    ]
}`)
	})

	var cmd = &Command{
		Args: []string{"list", "--service", "service", "--remote", "local", "--no-color"},
		Env:  []string{"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir:  "mocks/home/bucket/project",
	}

	cmd.Run()

	if update {
		tdata.ToFile("mocks/list/not-exists", cmd.Stderr.String())
	}

	var e = &Expect{
		Stderr:   tdata.FromFile("mocks/list/not-exists"),
		ExitCode: 1,
	}

	e.Assert(t, cmd)
}
