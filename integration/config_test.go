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
	e.AssertExact(t, cmd)
}

func TestConfigProjectList(t *testing.T) {
	var cmd = &Command{
		Args: []string{"config", "--list"},
		Dir:  "mocks/home/bucket/project",
	}

	var e = &Expect{
		Stdout: `id = app
name = my app
description = App example project
domain = app.liferay.io
`,
		ExitCode: 0,
	}

	cmd.Run()
	e.AssertExact(t, cmd)
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
	e.AssertExact(t, cmd)
}
