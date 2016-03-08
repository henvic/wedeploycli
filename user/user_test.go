package user

import "testing"

func TestHomeDir(t *testing.T) {
	var user = GetHomeDir()

	if len(user) == 0 {
		t.Errorf("Unexpected user home dir")
	}
}
