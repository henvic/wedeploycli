package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"strings"

	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/tdata"
)

func TestInspectPrintProjectStructure(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "project", "--fields"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var want = `ID string
CustomDomains []string
Health string
HomeService string
Description string`

	var e = &Expect{
		ExitCode: 0,
		Stdout:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestInspectPrintContainerStructure(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "project", "--fields"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var want = `ID string
		CustomDomains []string
		Health string
		HomeService string
		Description string`

	var e = &Expect{
		ExitCode: 0,
		Stdout:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestInspectPrintContextStructure(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "context", "--fields"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var want = `Scope usercontext.Scope
ProjectRoot string
ContainerRoot string
ProjectID string
ServiceID string
ProjectContainers []containers.ContainerInfo`

	var e = &Expect{
		ExitCode: 0,
		Stdout:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestInspectProjectFormat(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "project", "--format", "{{.ID}}"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/my-project/email",
	}

	var want = "my-project"

	var e = &Expect{
		ExitCode: 0,
		Stdout:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestInspectProjectFormatVerbose(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "project", "--format", "{{.ID}}", "--verbose"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/my-project/email",
	}

	cmd.Run()

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	if !strings.HasPrefix(cmd.Stderr.String(), "Reading project at") {
		t.Errorf("Expected output not found: %v", cmd.Stdout)
	}

	if !strings.Contains(cmd.Stderr.String(), `/cli/integration/mocks/inspect/my-project`) {
		t.Errorf("Expected err output not found: %v", cmd.Stderr)
	}

	if !strings.Contains(cmd.Stdout.String(), `my-project`) {
		t.Errorf("Expected output not found: %v", cmd.Stdout)
	}
}

func TestInspectProjectList(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "project"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/my-project/email",
	}

	cmd.Run()

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(cmd.Stdout.String()), &m); err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile("./mocks/inspect/my-project/expect.json"), m)

	if cmd.Stderr.Len() != 0 {
		t.Errorf("Wanted stderr to be empty, got %v instead", cmd.Stderr.String())
	}

	if cmd.ExitCode != 0 {
		t.Errorf("Wanted exit code to be 0, got %v instead", cmd.ExitCode)
	}
}

func TestInspectContainerFormat(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "container", "--format", "{{.Type}}"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/my-project/email",
	}

	var want = "wedeploy/email:latest"

	var e = &Expect{
		ExitCode: 0,
		Stdout:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestInspectContainerFormatVerbose(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "container", "--format", "{{.Type}}", "--verbose"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/my-project/email",
	}

	cmd.Run()

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	if !strings.HasPrefix(cmd.Stderr.String(), "Reading container at") {
		t.Errorf("Expected output not found: %v", cmd.Stdout)
	}

	if !strings.Contains(cmd.Stderr.String(), `/cli/integration/mocks/inspect/my-project/email`) {
		t.Errorf("Expected err output not found: %v", cmd.Stderr)
	}

	if !strings.Contains(cmd.Stdout.String(), `wedeploy/email:latest`) {
		t.Errorf("Expected output not found: %v", cmd.Stdout)
	}
}

func TestInspectContainerList(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "container"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/my-project/email",
	}

	cmd.Run()

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(cmd.Stdout.String()), &m); err != nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile("./mocks/inspect/my-project/email/expect.json"), m)

	if cmd.Stderr.Len() != 0 {
		t.Errorf("Wanted stderr to be empty, got %v instead", cmd.Stderr.String())
	}

	if cmd.ExitCode != 0 {
		t.Errorf("Wanted exit code to be 0, got %v instead", cmd.ExitCode)
	}
}

func TestInspectContextOnGlobalContextList(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "context"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks",
	}

	var want = tdata.FromFile("./mocks/inspect/context-list.json")

	var e = &Expect{
		ExitCode: 0,
		Stdout:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestInspectContextOnGlobalOnContainerOutsideProjectContextList(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "context"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/container-outside-project",
	}

	var want = fmt.Sprintf(tdata.FromFile("./mocks/inspect/container-outside-project/context-list.json"), abs("."))

	var e = &Expect{
		ExitCode: 0,
		Stdout:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestInspectContextOnProjectContextList(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "context"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/my-project",
	}

	var path = abs("./")
	var want = fmt.Sprintf(tdata.FromFile("./mocks/inspect/my-project/context-list.json"), path, path)

	var e = &Expect{
		ExitCode: 0,
		Stdout:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestInspectContextOnContainerContextList(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "context"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/my-project/email",
	}

	var path = abs("./")
	var want = fmt.Sprintf(tdata.FromFile("./mocks/inspect/my-project/email/context-list.json"), path, path, path)

	var e = &Expect{
		ExitCode: 0,
		Stdout:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestInspectContextListOnContainerOnProjectWithContainersInsideSubdirectories(t *testing.T) {
	defer Teardown()
	Setup()

	var mock = "./mocks/inspect/project-with-containers-inside-subdirs/sub-dir-2/sub-dir/container-level-3"

	var cmd = &Command{
		Args: []string{"inspect", "context"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: mock,
	}

	var path = abs("./mocks/inspect/project-with-containers-inside-subdirs")
	var want = strings.Replace(tdata.FromFile(mock+"/context-list.json"), "{{PATH}}", path, -1)

	var e = &Expect{
		ExitCode: 0,
		Stdout:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestInspectContextOnContainerContextFormat(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "context", "--format", "{{.ServiceID}}"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/my-project/email",
	}

	var want = "email\n"

	var e = &Expect{
		ExitCode: 0,
		Stdout:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
