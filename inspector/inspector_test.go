package inspector

import (
	"bytes"
	"testing"

	"encoding/json"

	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/stringlib"
	"github.com/wedeploy/cli/tdata"
)

var defaultOutStream = outStream

func TestPrintProjectSpec(t *testing.T) {
	var bufOutStream = bytes.Buffer{}
	outStream = &bufOutStream

	PrintProjectSpec()
	var want = `ID string
CustomDomains []string
Health string
Description string
Containers containers.Containers`

	stringlib.AssertSimilar(t, want, bufOutStream.String())

	outStream = defaultOutStream
}

func TestPrintContainerSpec(t *testing.T) {
	var bufOutStream = bytes.Buffer{}
	outStream = &bufOutStream

	PrintContainerSpec()
	var want = `ID string
Health string
Type string
Hooks *hooks.Hooks
Env map[string]string
Scale int`

	stringlib.AssertSimilar(t, want, bufOutStream.String())

	outStream = defaultOutStream
}

func TestInspectProjectList(t *testing.T) {
	var bufOutStream = bytes.Buffer{}
	outStream = &bufOutStream

	var err = InspectProject("", "./mocks/my-project")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var m map[string]interface{}
	if err = json.Unmarshal([]byte(bufOutStream.String()), &m); err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile("./mocks/my-project/expect.json"), m)

	outStream = defaultOutStream
}

func TestInspectProjectCustomDomain(t *testing.T) {
	var bufOutStream = bytes.Buffer{}
	outStream = &bufOutStream

	var err = InspectProject("{{(index .CustomDomains 0)}}", "./mocks/my-project")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = "example.net\n"
	var got = bufOutStream.String()

	if want != got {
		t.Errorf("Wanted custom domain to be %v, got %v instead", want, got)
	}

	outStream = defaultOutStream
}

func TestInspectProjectFormatError(t *testing.T) {
	var bufOutStream = bytes.Buffer{}
	outStream = &bufOutStream

	var err = InspectProject("{{.", "./mocks/my-project")
	var wantErr = `Template parsing error: template: :1: illegal number syntax: "."`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}

	var want = ""
	var got = bufOutStream.String()

	if want != got {
		t.Errorf("Wanted custom domain to be %v, got %v instead", want, got)
	}

	outStream = defaultOutStream
}

func TestInspectProjectNotFound(t *testing.T) {
	var err = InspectProject("", "./mocks/foo")
	var wantErr = `Inspection failure: can't find project`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectProjectCorrupted(t *testing.T) {
	var err = InspectProject("", "./mocks/corrupted-project")
	var wantErr = `Inspection failure on project: invalid character 'c' looking for beginning of object key string`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectContainerList(t *testing.T) {
	var bufOutStream = bytes.Buffer{}
	outStream = &bufOutStream

	var err = InspectContainer("", "./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var m map[string]interface{}
	if err = json.Unmarshal([]byte(bufOutStream.String()), &m); err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile("./mocks/my-project/email/expect.json"), m)

	outStream = defaultOutStream
}

func TestInspectContainerType(t *testing.T) {
	var bufOutStream = bytes.Buffer{}
	outStream = &bufOutStream

	var err = InspectContainer("{{.Type}}", "./mocks/my-project/email")

	if err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	var want = "wedeploy/email:latest\n"
	var got = bufOutStream.String()

	if want != got {
		t.Errorf("Wanted custom domain to be %v, got %v instead", want, got)
	}

	outStream = defaultOutStream
}

func TestInspectContainerFormatError(t *testing.T) {
	var bufOutStream = bytes.Buffer{}
	outStream = &bufOutStream

	var err = InspectContainer("{{.", "./mocks/my-project/email")
	var wantErr = `Template parsing error: template: :1: illegal number syntax: "."`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}

	var want = ""
	var got = bufOutStream.String()

	if want != got {
		t.Errorf("Wanted custom domain to be %v, got %v instead", want, got)
	}

	outStream = defaultOutStream
}

func TestInspectContainerNotFound(t *testing.T) {
	var err = InspectContainer("", "./mocks/my-project/container-not-found")
	var wantErr = `Inspection failure: can't find container`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestInspectContainerCorrupted(t *testing.T) {
	var err = InspectContainer("", "./mocks/my-project/container-corrupted")
	var wantErr = `Inspection failure on container: unexpected end of JSON input`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}
