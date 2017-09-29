package integration

import (
	"encoding/json"
	"testing"

	"strings"

	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/tdata"
)

func TestInspectPrintUnavailableStructure(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "unavailable"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	var want = `! Something went wrong with your "we inspect" operation.
! Inspecting "unavailable" is not implemented.`

	var e = &Expect{
		ExitCode: 1,
		Stderr:   want,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestInspectPrintServiceStructure(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "service", "--fields"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
	}

	cmd.Run()

	if update {
		tdata.ToFile("mocks/inspect/print-service-structure", cmd.Stdout.String())
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/inspect/print-service-structure"),
	}

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

	cmd.Run()

	if update {
		tdata.ToFile("mocks/inspect/print-context-structure", cmd.Stdout.String())
	}

	var e = &Expect{
		ExitCode: 0,
		Stdout:   tdata.FromFile("mocks/inspect/print-context-structure"),
	}

	e.Assert(t, cmd)
}

func TestInspectServiceFormat(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "service", "--format", "{{.Image}}"},
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

func TestInspectServiceFormatVerbose(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "service", "--format", "{{.Image}}", "--verbose"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/my-project/email",
	}

	cmd.Run()

	if cmd.ExitCode != 0 {
		t.Errorf("Expected exit code to be 0, got %v instead", cmd.ExitCode)
	}

	if !strings.HasPrefix(cmd.Stderr.String(), "Reading service at") {
		t.Errorf("Expected output not found: %v", cmd.Stdout)
	}

	if !strings.Contains(cmd.Stderr.String(), `/cli/integration/mocks/inspect/my-project/email`) {
		t.Errorf("Expected err output not found: %v", cmd.Stderr)
	}

	if !strings.Contains(cmd.Stdout.String(), `wedeploy/email:latest`) {
		t.Errorf("Expected output not found: %v", cmd.Stdout)
	}
}

func TestInspectServiceList(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "service"},
		Env: []string{
			"WEDEPLOY_CUSTOM_HOME=" + GetLoginHome()},
		Dir: "./mocks/inspect/my-project/email",
	}

	cmd.Run()

	var m map[string]interface{}
	if err := json.Unmarshal(cmd.Stdout.Bytes(), &m); err != nil {
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

func TestInspectContextContextList(t *testing.T) {
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

func TestInspectContextOnServiceContextFormat(t *testing.T) {
	defer Teardown()
	Setup()

	var cmd = &Command{
		Args: []string{"inspect", "context", "--format", "{{(index .Services 0).ServiceID}}"},
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
