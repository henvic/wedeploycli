package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

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
	e.Assert(t, cmd)
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
	e.Assert(t, cmd)
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
	e.Assert(t, cmd)
}

func TestCorruptConfig(t *testing.T) {
	var cmd = &Command{
		Args: []string{"projects", "-v"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetRegularHome()},
		Dir:  "mocks/home/bucket/invalid-project",
	}

	var p, err = os.Getwd()

	if err != nil {
		panic(err)
	}

	p = filepath.Join(p, "mocks/home/bucket/invalid-project/project.json")

	var e = &Expect{
		Stderr: fmt.Sprintf(`Unexpected error reading configuration file.
Fix %v by hand or erase it.
`, p),
		ExitCode: 1,
	}

	cmd.Run()
	e.Assert(t, cmd)
}

func TestLoggedOut(t *testing.T) {
	var cmd = &Command{
		Args: []string{"projects", "-v"},
		Env:  []string{"LAUNCHPAD_CUSTOM_HOME=" + GetLogoutHome()},
	}

	var e = &Expect{
		Stderr:   "Please run \"launchpad login\" first.\n",
		ExitCode: 1,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
