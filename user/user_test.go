package user

import (
	"os"
	"testing"
)

func TestHomeDir(t *testing.T) {
	var user = GetHomeDir()

	if len(user) == 0 {
		t.Errorf("Unexpected user home dir")
	}
}

func TestCustomHomeDir(t *testing.T) {
	var want = "foo"
	var defaultEnv = os.Getenv("LAUNCHPAD_CUSTOM_HOME")

	if err := os.Setenv("LAUNCHPAD_CUSTOM_HOME", "foo"); err != nil {
		panic(err)
	}

	var got = GetHomeDir()

	if got != want {
		t.Errorf("Wanted custom home to be %v, got %v instead", want, got)
	}

	if err := os.Setenv("LAUNCHPAD_CUSTOM_HOME", defaultEnv); err != nil {
		panic(err)
	}
}
