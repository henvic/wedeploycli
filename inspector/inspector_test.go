package inspector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/projects"
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
	var want = []string{`ID string`,
		`CustomDomains []string`,
		`Health string`,
		`Description string`,
		`Containers containers.Containers`}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted spec %v, got %v instead", want, got)
	}
}

func TestPrintContainerSpec(t *testing.T) {
	var got = GetSpec(containers.Container{})
	var want = []string{`ID string`,
		`Health string`,
		`Type string`,
		`Hooks *hooks.Hooks`,
		`Env map[string]string`,
		`Scale int`}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted spec %v, got %v instead", want, got)
	}
}

func TestPrintContextSpec(t *testing.T) {
	var got = GetSpec(ContextOverview{})
	var want = []string{`Scope usercontext.Scope`,
		`ProjectRoot string`,
		`ContainerRoot string`,
		`ProjectID string`,
		`ContainerID string`}

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

	jsonlib.AssertJSONMarshal(t, tdata.FromFile("./mocks/my-project/expect.json"), m)
}

func TestInspectProjectCustomDomain(t *testing.T) {
	var got, err = InspectProject("{{(index .CustomDomains 0)}}", "./mocks/my-project")

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
	var wantErr = `Inspection failure: can't find project`

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
	var wantErr = fmt.Sprintf(`Can't load project context on %v:`+
		` Inspection failure on project: invalid character 'c' looking for beginning of object key string`,
		abs("mocks/corrupted-project"))

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectContainerList(t *testing.T) {
	var got, err = InspectContainer("", "./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var m map[string]interface{}
	if err = json.Unmarshal([]byte(got), &m); err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile("./mocks/my-project/email/expect.json"), m)
}

func TestInspectContainerType(t *testing.T) {
	var got, err = InspectContainer("{{.Type}}", "./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = "wedeploy/email:latest"

	if want != got {
		t.Errorf("Wanted custom domain to be %v, got %v instead", want, got)
	}
}

func TestInspectContainerFormatError(t *testing.T) {
	var got, err = InspectContainer("{{.", "./mocks/my-project/email")
	var wantErr = `Template parsing error: template: :1: illegal number syntax: "."`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}

	var want = ""

	if want != got {
		t.Errorf("Wanted custom domain to be %v, got %v instead", want, got)
	}
}

func TestInspectContainerNotFound(t *testing.T) {
	var _, err = InspectContainer("", "./mocks/my-project/container-not-found")
	var wantErr = `Inspection failure: can't find container`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectContainerCorrupted(t *testing.T) {
	var _, err = InspectContainer("", "./mocks/my-project/corrupted-container")
	var wantErr = `Inspection failure on container: unexpected end of JSON input`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectContainerCorruptedOnContextOverview(t *testing.T) {
	var _, err = InspectContext("", "./mocks/my-project/corrupted-container")
	var wantErr = fmt.Sprintf(`Can't load container context on %v:`+
		` Inspection failure on container: unexpected end of JSON input`, abs("mocks/my-project/corrupted-container"))

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
		Scope:         "global",
		ProjectRoot:   "",
		ContainerRoot: "",
		ProjectID:     "",
		ContainerID:   "",
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
		Scope:         "project",
		ProjectRoot:   abs("./mocks/my-project"),
		ContainerRoot: "",
		ProjectID:     "my-project",
		ContainerID:   "",
	}

	if !reflect.DeepEqual(want, overview) {
		t.Errorf("Wanted ContextOverview %+v, got %+v instead", want, overview)
	}
}

func TestInspectContextOverviewTypeContainer(t *testing.T) {
	var overview = ContextOverview{}
	var err = overview.Load("./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = ContextOverview{
		Scope:         "container",
		ProjectRoot:   abs("./mocks/my-project"),
		ContainerRoot: abs("./mocks/my-project/email"),
		ProjectID:     "my-project",
		ContainerID:   "email",
	}

	if !reflect.DeepEqual(want, overview) {
		t.Errorf("Wanted ContextOverview %+v, got %+v instead", want, overview)
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
    "ContainerRoot": "",
    "ProjectID": "my-project",
    "ContainerID": ""
}`,
		abs("mocks/my-project"))

	if got != want {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestInspectContextOverviewContainer(t *testing.T) {
	var got, err = InspectContext("", "./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = fmt.Sprintf(`{
    "Scope": "container",
    "ProjectRoot": "%v",
    "ContainerRoot": "%v",
    "ProjectID": "my-project",
    "ContainerID": "email"
}`,
		abs("mocks/my-project"),
		abs("mocks/my-project/email"))

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
