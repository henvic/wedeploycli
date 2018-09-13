package deployremote

import (
	"os"
	"strings"
	"testing"

	"github.com/wedeploy/cli/userhome"
)

func TestGetWorkingDirectory(t *testing.T) {
	var perm, _ = os.Getwd()

	defer func() {
		if err := os.Chdir(perm); err != nil {
			panic(err)
		}
	}()

	// before changing to blacklisted
	var got, err = getWorkingDirectory()

	if err != nil {
		t.Errorf("Can't get working directory: %v", err)
	}

	if err = os.Chdir(userhome.GetHomeDir()); err != nil {
		panic(err)
	}

	if got != perm {
		t.Errorf("Expected value to be %v instead of %v", perm, got)
	}

	// after blacklisting...
	got, err = getWorkingDirectory()

	if got != "" {
		t.Errorf("Expected working directory to be empty")
	}

	const refuse = "refusing to deploy"

	if err == nil || !strings.Contains(err.Error(), refuse) {
		t.Errorf("Expected error to contain %v, got %v instead", refuse, err)
	}
}
