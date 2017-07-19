package inspector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/tdata"

	"reflect"

	"github.com/wedeploy/cli/findresource"
)

func TestMain(m *testing.M) {
	var (
		defaultSysRoot string
		ec             int
	)

	defer func() {
		if err := findresource.SetSysRoot(defaultSysRoot); err != nil {
			panic(err)
		}

		os.Exit(ec)
	}()

	defaultSysRoot = findresource.GetSysRoot()

	if err := findresource.SetSysRoot("./mocks"); err != nil {
		panic(err)
	}

	ec = m.Run()
}

func TestGetSpecContextOverview(t *testing.T) {
	var got = GetSpec(projects.Project{})
	var want = []string{`ProjectID string`,
		`Health string`,
		`Description string`,
		`HealthUID string`}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted spec %v, got %v instead", want, got)
	}
}

func TestPrintServiceSpec(t *testing.T) {
	var got = GetSpec(services.ServicePackage{})
	var want = []string{`ID string`,
		`Scale int`,
		`Type string`,
		`Hooks *hooks.Hooks`,
		`CustomDomains []string`,
		`Env map[string]string`,
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted spec %v, got %v instead", want, got)
	}
}

func TestPrintContextSpec(t *testing.T) {
	var got = GetSpec(ContextOverview{})
	var want = []string{`Scope usercontext.Scope`,
		`ProjectRoot string`,
		`ServiceRoot string`,
		`ProjectID string`,
		`ServiceID string`,
		`ProjectServices []services.ServiceInfo`}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted spec %v, got %v instead", want, got)
	}
}

func TestInspectProjectList(t *testing.T) {
	var got, err = InspectProject("", "./mocks/my-project")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var m map[string]interface{}
	if err = json.Unmarshal([]byte(got), &m); err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, `{"id": "my-project"}`, m)
}

func TestInspectServiceCustomDomain(t *testing.T) {
	var got, err = InspectService("{{(index .CustomDomains 0)}}", "./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = "example.net"

	if want != got {
		t.Errorf("Wanted custom domain to be %v, got %v instead", want, got)
	}
}

func TestInspectProjectFormatError(t *testing.T) {
	var got, err = InspectProject("{{.", "./mocks/my-project")
	var wantErr = `Template parsing error: template: :1: illegal number syntax: "."`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}

	var want = ""

	if want != got {
		t.Errorf("Wanted custom domain to be %v, got %v instead", want, got)
	}
}

func TestInspectProjectNotFound(t *testing.T) {
	var _, err = InspectProject("", "./mocks/foo")
	var wantErr = `Inspection failure: can not find project`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectProjectCorrupted(t *testing.T) {
	var _, err = InspectProject("", "./mocks/corrupted-project")
	var wantErr = `Inspection failure on project: invalid character 'c' looking for beginning of object key string`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectProjectCorruptedOnContextOverview(t *testing.T) {
	var _, err = InspectContext("", "./mocks/corrupted-project")
	var wantErr = fmt.Sprintf(`Can not load project context on %v:`+
		` Inspection failure on project: invalid character 'c' looking for beginning of object key string`,
		abs("mocks/corrupted-project"))

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectServiceList(t *testing.T) {
	var got, err = InspectService("", "./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var m map[string]interface{}
	if err = json.Unmarshal([]byte(got), &m); err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile("./mocks/my-project/email/expect.json"), m)
}

func TestInspectServiceType(t *testing.T) {
	var got, err = InspectService("{{.Type}}", "./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = "wedeploy/email:latest"

	if want != got {
		t.Errorf("Wanted custom domain to be %v, got %v instead", want, got)
	}
}

func TestInspectServiceFormatError(t *testing.T) {
	var got, err = InspectService("{{.", "./mocks/my-project/email")
	var wantErr = `Template parsing error: template: :1: illegal number syntax: "."`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}

	var want = ""

	if want != got {
		t.Errorf("Wanted custom domain to be %v, got %v instead", want, got)
	}
}

func TestInspectServiceNotFound(t *testing.T) {
	var _, err = InspectService("", "./mocks/my-project/service-not-found")
	var wantErr = `Inspection failure: can not find service`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectServiceCorrupted(t *testing.T) {
	var _, err = InspectService("", "./mocks/project-with-corrupted-service/corrupted-service")
	var wantErr = `Inspection failure on service: unexpected end of JSON input`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectProjectWithCorruptedServiceOnContextOverview(t *testing.T) {
	var _, err = InspectContext("", "./mocks/project-with-corrupted-service")
	var wantErr = fmt.Sprintf(`Error while trying to read list of services on project: `+
		`Can not list services: error reading %v: unexpected end of JSON input`,
		abs("mocks/project-with-corrupted-service/corrupted-service/wedeploy.json"))

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectContextOverviewTypeGlobal(t *testing.T) {
	var overview = ContextOverview{}
	var err = overview.Load("./mocks/")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = ContextOverview{
		Scope:       "global",
		ProjectRoot: "",
		ServiceRoot: "",
		ProjectID:   "",
		ServiceID:   "",
	}

	if !reflect.DeepEqual(want, overview) {
		t.Errorf("Wanted ContextOverview %+v, got %+v instead", want, overview)
	}
}

func TestInspectContextOverviewTypeProject(t *testing.T) {
	var overview = ContextOverview{}
	var err = overview.Load("./mocks/my-project/")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = ContextOverview{
		Scope:       "project",
		ProjectRoot: abs("./mocks/my-project"),
		ServiceRoot: "",
		ProjectID:   "my-project",
		ServiceID:   "",
		ProjectServices: services.ServiceInfoList{
			services.ServiceInfo{
				ServiceID: "email",
				Location:  abs("./mocks/my-project/email"),
			},
			services.ServiceInfo{
				ServiceID: "other",
				Location:  abs("./mocks/my-project/other"),
			},
		},
	}

	if !reflect.DeepEqual(want, overview) {
		t.Errorf("Wanted ContextOverview %+v, got %+v instead", want, overview)
	}
}

func TestInspectContextOverviewTypeProjectWithDuplicatedServices(t *testing.T) {
	var overview = ContextOverview{}
	var err = overview.Load("./mocks/my-project-with-duplicated-services/")

	if err == nil || strings.Contains(err.Error(),
		"Error while trying to read list of services on project:\n"+
			`Can not list services: ID "other" was found duplicated on services`) {
		t.Errorf("Expected error to contain duplicated information, got %v instead", err)
	}

	var want = ContextOverview{
		Scope:       "project",
		ProjectRoot: abs("./mocks/my-project-with-duplicated-services"),
		ServiceRoot: "",
		ProjectID:   "my-project",
		ServiceID:   "",
	}

	if !reflect.DeepEqual(want, overview) {
		t.Errorf("Wanted ContextOverview %+v, got %+v instead", want, overview)
	}
}

func TestInspectContextOverviewTypeService(t *testing.T) {
	var overview = ContextOverview{}
	var err = overview.Load("./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = ContextOverview{
		Scope:       "service",
		ProjectRoot: abs("./mocks/my-project"),
		ServiceRoot: abs("./mocks/my-project/email"),
		ProjectID:   "my-project",
		ServiceID:   "email",
		ProjectServices: services.ServiceInfoList{
			services.ServiceInfo{
				ServiceID: "email",
				Location:  abs("./mocks/my-project/email"),
			},
			services.ServiceInfo{
				ServiceID: "other",
				Location:  abs("./mocks/my-project/other"),
			},
		},
	}

	if !reflect.DeepEqual(want, overview) {
		t.Errorf("Wanted ContextOverview %+v, got %+v instead", want, overview)
	}
}

func TestInspectContextOverviewTypeServiceOutsideProject(t *testing.T) {
	var overview = ContextOverview{}
	var err = overview.Load("./mocks/service-outside-project")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = ContextOverview{
		Scope:           "global",
		ServiceRoot:     abs("./mocks/service-outside-project"),
		ProjectID:       "",
		ServiceID:       "alone",
		ProjectServices: nil,
	}

	if !reflect.DeepEqual(want, overview) {
		t.Errorf("Wanted ContextOverview %+v, got %+v instead", want, overview)
	}
}

func TestInspectServiceCorruptedOnContextOverview(t *testing.T) {
	var _, err = InspectContext("", "./mocks/corrupted-service-outside-project")
	var wantErr = fmt.Sprintf(`Can not load service context on %v: `+
		`Inspection failure on service: invalid character ':' after top-level value`,
		abs("./mocks/corrupted-service-outside-project"))

	if err == nil || err.Error() != wantErr {
		t.Errorf("Error should be %v, got %v instead", wantErr, err)
	}
}

func TestInspectContextOverviewProject(t *testing.T) {
	var got, err = InspectContext("", "./mocks/my-project")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = fmt.Sprintf(`{
    "Scope": "project",
    "ProjectRoot": "%v",
    "ServiceRoot": "",
    "ProjectID": "my-project",
    "ServiceID": "",
    "ProjectServices": [
        {
            "ServiceID": "email",
            "Location": "%v"
        },
        {
            "ServiceID": "other",
            "Location": "%v"
        }
    ]
}`,
		abs("mocks/my-project"),
		abs("mocks/my-project/email"),
		abs("mocks/my-project/other"))

	if got != want {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestInspectContextOverviewService(t *testing.T) {
	var got, err = InspectContext("", "./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = fmt.Sprintf(`{
    "Scope": "service",
    "ProjectRoot": "%v",
    "ServiceRoot": "%v",
    "ProjectID": "my-project",
    "ServiceID": "email",
    "ProjectServices": [
        {
            "ServiceID": "email",
            "Location": "%v"
        },
        {
            "ServiceID": "other",
            "Location": "%v"
        }
    ]
}`,
		abs("mocks/my-project"),
		abs("mocks/my-project/email"),
		abs("mocks/my-project/email"),
		abs("mocks/my-project/other"))

	if got != want {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func abs(path string) string {
	var abs, err = filepath.Abs(path)

	if err != nil {
		panic(err)
	}

	return abs
}
