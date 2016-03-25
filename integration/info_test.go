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
	e.AssertExact(t, cmd)
}

func TestInfoProject(t *testing.T) {
	var cmd = &Command{
		Args: []string{"info"},
		Dir:  "mocks/home/bucket/project",
	}

	var e = &Expect{
		Stdout: `Project: app (my app)
Domain: app.liferay.io
Description: App example project
`,
		ExitCode: 0,
	}

	cmd.Run()
	e.AssertExact(t, cmd)
}

func TestInfoContainer(t *testing.T) {
	var cmd = &Command{
		Args: []string{"info"},
		Dir:  "mocks/home/bucket/project/container",
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
	e.AssertExact(t, cmd)
}
