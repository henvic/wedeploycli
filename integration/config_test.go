package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

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
		Stderr:   "Please run \"we login\" first.\n",
		ExitCode: 1,
	}

	cmd.Run()
	e.Assert(t, cmd)
}
