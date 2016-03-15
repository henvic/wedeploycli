package globalconfigmock

import (
	"testing"

	"github.com/launchpad-project/cli/config"
)

func TestGlobalConfigMock(t *testing.T) {
	if len(config.Stores) != 0 {
		t.Error("Expected config.Store to have no config")
	}

	Setup()

	if config.Stores["global"] == nil {
		t.Error("Expected global config to be mocked")
	}

	var global = config.Stores["global"]

	var username, err = global.GetString("username")

	if username != "admin" || err != nil {
		t.Error("Failed to retrieve expected config data from the mock")
	}

	Teardown()

	if config.Stores["global"] != nil {
		t.Error("Expected global config to be unmocked")
	}
}
