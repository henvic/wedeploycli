package globalconfigmock

import (
	"os"

	"github.com/launchpad-project/cli/config"
)

var original *config.Config

// Setup the global config mock
func Setup() {
	original = config.Global

	var mock = &config.Config{
		Path: os.DevNull,
	}

	mock.Load()
	setMockDefaults(mock)
	config.Global = mock
}

// Teardown the global config mock
func Teardown() {
	config.Global = original
}

func setMockDefaults(mock *config.Config) {
	mock.Username = "admin"
	mock.Password = "safe"
	mock.Local = false
	mock.NoColor = false
	mock.Endpoint = "http://www.example.com"
	mock.NotifyUpdates = true
	mock.ReleaseChannel = "stable"
	mock.LastUpdateCheck = "Sat Jun  4 04:47:03 BRT 2016"
}
