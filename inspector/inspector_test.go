package inspector

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/services"
	"github.com/wedeploy/cli/tdata"
	"github.com/wedeploy/wedeploy-sdk-go/jsonlib"

	"reflect"

	"github.com/wedeploy/cli/findresource"
)

var update bool

func init() {
	flag.BoolVar(&update, "update", false, "update golden files")
}

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

func TestPrintServiceSpec(t *testing.T) {
	var got = GetSpec(services.Package{})
	var want = []string{
		`ID string`,
		`ProjectID string`,
		`Scale int`,
		`Image string`,
		`CustomDomains []string`,
		`Env map[string]string`,
		`Dependencies []string`,
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted spec %v, got %v instead", want, got)
	}
}

func TestPrintContextSpec(t *testing.T) {
	var got = GetSpec(ContextOverview{})
	var want = []string{
		`ProjectID string`,
		`Services []services.ServiceInfo`,
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Wanted spec %v, got %v instead", want, got)
	}
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

func TestInspectServiceList(t *testing.T) {
	var got, err = InspectService("", "./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var m map[string]interface{}
	if err = json.Unmarshal([]byte(got), &m); err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	if update {
		var b, err = json.Marshal(m)

		if err != nil {
			panic(err)
		}

		tdata.ToFile("./mocks/my-project/email/expect.json", string(b))
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile("./mocks/my-project/email/expect.json"), m)
}

func TestInspectServiceImage(t *testing.T) {
	var got, err = InspectService("{{.Image}}", "./mocks/my-project/email")

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
	var wantErr = `template parsing error: template: :1: illegal number syntax: "."`

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
	var wantErr = `inspection failure: can't find service`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectServiceCorrupted(t *testing.T) {
	var _, err = InspectService("", "./mocks/project-with-corrupted-service/corrupted-service")
	var wantErr = fmt.Sprintf(`error parsing wedeploy.json on %v: unexpected end of JSON input`,
		abs("./mocks/project-with-corrupted-service/corrupted-service"))

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectProjectWithCorruptedServiceOnContextOverview(t *testing.T) {
	var _, err = InspectContext("", "./mocks/project-with-corrupted-service")
	var wantErr = fmt.Sprintf(
		`error parsing wedeploy.json on %v: unexpected end of JSON input`,
		abs("mocks/project-with-corrupted-service/corrupted-service"))

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectContextOverviewWithDuplicatedServices(t *testing.T) {
	var overview = ContextOverview{}
	var err = overview.Load("./mocks/my-project-with-duplicated-services/")

	if err == nil || strings.Contains(err.Error(),
		"Error while trying to read list of services on project:\n"+
			`ID "other" was found duplicated on services`) {
		t.Errorf("Expected error to contain duplicated information, got %v instead", err)
	}
}

func TestInspectServiceCorruptedOnContextOverview(t *testing.T) {
	var _, err = InspectContext("", "./mocks/corrupted-service-outside-project")
	var wantErr = fmt.Sprintf(`error parsing wedeploy.json on %v:`+
		` invalid character ':' after top-level value.
The wedeploy.json file syntax is described at https://help.liferay.com/hc/en-us/articles/360012918551-Configuring-via-the-wedeploy-json`,
		abs("./mocks/corrupted-service-outside-project"))

	if err == nil || err.Error() != wantErr {
		t.Errorf("Error should be %v, got %v instead", wantErr, err)
	}
}

func TestInspectContextOverview(t *testing.T) {
	var got, err = InspectContext("", "./mocks/my-project")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = fmt.Sprintf(`{
    "ProjectID": "exampleProject",
    "Services": [
        {
            "ProjectID": "exampleProject",
            "ServiceID": "email",
            "Location": "%v"
        },
        {
            "ProjectID": "exampleProject",
            "ServiceID": "other",
            "Location": "%v"
        }
    ]
}`,
		abs("mocks/my-project/email"),
		abs("mocks/my-project/other"))

	if got != want {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestInspectContextOverviewMismatchedProjectID(t *testing.T) {
	var _, err = InspectContext("", "./mocks/project-with-mismatched-project-id")

	var wantErr = `services "email" and "other" must have the same project ID defined on "email/wedeploy.json" and "other/wedeploy.json" (currently: "exampleProject" and "notExampleProject")`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectContextOverviewMismatchedProjectIDWhenEmpty(t *testing.T) {
	var _, err = InspectContext("", "./mocks/project-with-mismatched-project-id-2")

	var wantErr = `services "email" and "other" must have the same project ID defined on "email/wedeploy.json" and "other/wedeploy.json" (currently: "" and "notExampleProject")`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectContextOverviewService(t *testing.T) {
	var got, err = InspectContext("", "./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = fmt.Sprintf(`{
    "ProjectID": "exampleProject",
    "Services": [
        {
            "ProjectID": "exampleProject",
            "ServiceID": "email",
            "Location": "%v"
        }
    ]
}`,
		abs("mocks/my-project/email"))

	if got != want {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestInspectConfig(t *testing.T) {
	var params = config.Params{
		ReleaseChannel: "testing",
		NextVersion:    "3.0",
	}

	wectx := config.NewContext(config.ContextParams{})

	wectx.Config().SetParams(params)

	var got, err = InspectConfig("", wectx)

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	var want = `{
    "DefaultRemote": "",
    "NoAutocomplete": false,
    "NoColor": false,
    "NotifyUpdates": false,
    "ReleaseChannel": "testing",
    "LastUpdateCheck": "",
    "PastVersion": "",
    "NextVersion": "3.0",
    "EnableAnalytics": false,
    "AnalyticsID": "",
    "EnableCURL": false,
    "Remotes": null
}`

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
