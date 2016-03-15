package globalconfigmock

import (
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/configstore"
)

var originalCSG *configstore.Store

// Setup the global config mock
func Setup() {
	originalCSG = config.Stores["global"]

	var mockGlobal = configstore.Store{
		Name: "global",
		Path: "../apihelper/mocks/config.json",
	}

	mockGlobal.Load()
	config.Stores["global"] = &mockGlobal
}

// Teardown the global config mock
func Teardown() {
	config.Stores["global"] = originalCSG
}
