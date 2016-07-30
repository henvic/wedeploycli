package configmock

import (
	"fmt"
	"os"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/context"
)

var originalGlobal *config.Config
var originalContext *context.Context

// Setup the global config mock
func Setup() {
	originalGlobal = config.Global
	originalContext = config.Context

	var global = &config.Config{
		Path: os.DevNull,
	}

	global.Load()
	global.Username = "admin"
	global.Password = "safe"
	global.Token = "mock_token"
	global.Local = false
	global.NoColor = false
	global.Endpoint = "http://www.example.com"
	global.NotifyUpdates = true
	global.ReleaseChannel = "stable"
	global.LastUpdateCheck = "Sat Jun  4 04:47:03 BRT 2016"
	config.Global = global

	config.Context = &context.Context{
		Username: global.Username,
		Password: global.Password,
		Token:    global.Token,
		Endpoint: fmt.Sprintf("http://localhost:8080/"),
	}

	SetupLocalContext()
}

// SetupLocalContext loads the config mock local context
func SetupLocalContext() {
	config.Context.Remote = ""
	config.Context.Endpoint = "http://localhost:8080/"
	config.Context.Token = "1"
}

// SetupRemoteContext loads the config mock remote context
func SetupRemoteContext() {
	config.Context.Remote = "foo"
	config.Context.Endpoint = config.Global.Endpoint
	config.Context.Token = config.Global.Token
}

// Teardown the global config mock
func Teardown() {
	config.Global = originalGlobal
	config.Context = originalContext
}
