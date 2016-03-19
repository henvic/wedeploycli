package integration

import "testing"

func TestConfigGlobalList(t *testing.T) {
	var cmd = &Command{
		Args: []string{"config", "--list"},
	}

	var e = &Expect{
		Stdout: `username = admin
password = safe
endpoint = http://www.example.com/
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
		t.Errorf("Wanted Stderr %v, got %v instead", e.Stderr, errString)
	}

	if outString != e.Stdout {
		t.Errorf("Wanted Stdout %v, got %v instead", e.Stdout, outString)
	}
}

func TestConfigProjectList(t *testing.T) {
	var cmd = &Command{
		Args: []string{"config", "--list"},
		Dir:  "mocks/home/bucket/project",
	}

	var e = &Expect{
		Stdout: `name = app
description = App example project
domain = app.liferay.io
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
		t.Errorf("Wanted Stderr %v, got %v instead", e.Stderr, errString)
	}

	if outString != e.Stdout {
		t.Errorf("Wanted Stdout %v, got %v instead", e.Stdout, outString)
	}
}

func TestConfigContainerList(t *testing.T) {
	var cmd = &Command{
		Args: []string{"config", "--list"},
		Dir:  "mocks/home/bucket/project/container",
	}

	var e = &Expect{
		ExitCode: 0,
	}

	cmd.Run()

	if cmd.ExitCode != e.ExitCode {
		t.Errorf("Wanted exit code %v, got %v instead", e.ExitCode, cmd.ExitCode)
	}

	errString := cmd.Stderr.String()
	outString := cmd.Stdout.String()

	if errString != e.Stderr {
		t.Errorf("Wanted Stderr %v, got %v instead", e.Stderr, errString)
	}

	if outString != e.Stdout {
		t.Errorf("Wanted Stdout %v, got %v instead", e.Stdout, outString)
	}
}
