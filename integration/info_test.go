package integration

import "testing"

func TestInfo(t *testing.T) {
	var cmd = &Command{
		Args: []string{"info"},
	}

	var e = &Expect{
		Stderr:   "fatal: not a project\n",
		ExitCode: 1,
	}

	cmd.Run()

	if cmd.ExitCode != e.ExitCode {
		t.Errorf("Wanted exit code %v, got %v instead", e.ExitCode, cmd.ExitCode)
	}

	errString := cmd.Stderr.String()
	outString := cmd.Stdout.String()

	if errString != e.Stderr {
		t.Errorf("Wanted Stderr %v, got %v instead", errString, e.Stderr)
	}

	if outString != e.Stdout {
		t.Errorf("Wanted Stdout %v, got %v instead", outString, e.Stdout)
	}
}

func TestInfoProject(t *testing.T) {
	var cmd = &Command{
		Args: []string{"info"},
		Dir:  "homemock/bucket/project",
	}

	var e = &Expect{
		Stdout: `Project: app
Domain: app.liferay.io
Description: App example project
`,
		ExitCode: 0,
	}

	cmd.Run()

	if cmd.ExitCode != e.ExitCode {
		t.Errorf("Wanted exit code %v, got %v instead", e.ExitCode, cmd.ExitCode)
	}

	errString := cmd.Stderr.String()
	outString := cmd.Stdout.String()

	if errString != e.Stderr {
		t.Errorf("Wanted Stderr %v, got %v instead", errString, e.Stderr)
	}

	if outString != e.Stdout {
		t.Errorf("Wanted Stdout %v, got %v instead", outString, e.Stdout)
	}
}

func TestInfoContainer(t *testing.T) {
	var cmd = &Command{
		Args: []string{"info"},
		Dir:  "homemock/bucket/project/container",
	}

	var e = &Expect{
		Stdout: `Container: 
Description: Static hosting container example
Version: 0.0.1
Runtime: static
`,
		ExitCode: 0,
	}

	cmd.Run()

	if cmd.ExitCode != e.ExitCode {
		t.Errorf("Wanted exit code %v, got %v instead", e.ExitCode, cmd.ExitCode)
	}

	errString := cmd.Stderr.String()
	outString := cmd.Stdout.String()

	if errString != e.Stderr {
		t.Errorf("Wanted Stderr %v, got %v instead", errString, e.Stderr)
	}

	if outString != e.Stdout {
		t.Errorf("Wanted Stdout %v, got %v instead", outString, e.Stdout)
	}
}
