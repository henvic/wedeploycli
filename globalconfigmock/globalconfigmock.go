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
		Data: map[string]interface{}{
			"username": "admin",
			"password": "safe",
			"endpoint": "http://www.example.com/",
		},
	}

	mockGlobal.Load()
	config.Stores["global"] = &mockGlobal
}

// Teardown the global config mock
func Teardown() {
	config.Stores["global"] = originalCSG
}
