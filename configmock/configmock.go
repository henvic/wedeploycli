package configmock

import (
	"fmt"
	"os"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/usercontext"
)

var originalGlobal *config.Config
var originalContext *usercontext.Context

// Setup the global config mock
func Setup() {
	originalGlobal = config.Global
	originalContext = config.Context

	var global = &config.Config{
		Path: os.DevNull,
	}

	if err := global.Load(); err != nil {
		panic(err)
	}

	global.Username = "admin"
	global.Password = "safe"
	global.Token = "mock_token"
	global.NoColor = false
	global.NotifyUpdates = true
	global.ReleaseChannel = "stable"
	global.LastUpdateCheck = "Sat Jun  4 04:47:03 -0300 2016"
	config.Global = global

	config.Context = &usercontext.Context{
		Scope:    usercontext.GlobalScope,
		Username: global.Username,
		Password: global.Password,
		Token:    global.Token,
		Endpoint: fmt.Sprintf("http://localhost:3002/"),
	}

	SetupLocalContext()
}

// SetupLocalContext loads the config mock local context
func SetupLocalContext() {
	config.Context.Remote = ""
	config.Context.RemoteAddress = "wedeploy.me"
	config.Context.Endpoint = "http://localhost:3002/"
	config.Context.Username = "no-reply@wedeploy.com"
	config.Context.Password = "cli-tool-password"
	config.Context.Token = ""
}

// SetupRemoteContext loads the config mock remote context
func SetupRemoteContext() {
	config.Context.Remote = "foo"
	config.Context.Endpoint = "http://www.example.com/"
	config.Context.Username = ""
	config.Context.Password = ""
	config.Context.Token = config.Global.Token
}

// Teardown the global config mock
func Teardown() {
	config.Global = originalGlobal
	config.Context = originalContext
}
