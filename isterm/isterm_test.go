package isterm

import (
	"os"
	"testing"

	"github.com/wedeploy/cli/envs"
)

func TestCheckForced(t *testing.T) {
	os.Clearenv()

	_ = os.Setenv(envs.SkipTerminalVerification, "true")

	if !Check() {
		t.Errorf("Check expected to be forced")
	}
}
