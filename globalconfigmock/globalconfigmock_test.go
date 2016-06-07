package globalconfigmock

import (
	"testing"

	"github.com/wedeploy/cli/config"
)

func TestGlobalConfigMock(t *testing.T) {
	Setup()

	if config.Global == nil {
		t.Error("Expected global config to be mocked")
	}

	if config.Global.Username != "admin" {
		t.Error("Failed to retrieve expected config data from the mock")
	}

	Teardown()

	if config.Global != original {
		t.Error("Expected global config to be unmocked")
	}
}
