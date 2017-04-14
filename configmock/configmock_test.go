package configmock

import (
	"testing"

	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
)

func TestConfigMock(t *testing.T) {
	Setup()

	if config.Global == nil {
		t.Error("Expected global config to be mocked")
	}

	if config.Global.Username != "admin" {
		t.Error("Failed to retrieve expected config data from the mock")
	}

	Teardown()

	if config.Global != originalGlobal {
		t.Error("Expected global config to be unmocked")
	}

	if config.Context != originalContext {
		t.Error("Expected context to be unmocked")
	}
}

func TestConfigMockDefaultContext(t *testing.T) {
	Setup()

	if config.Context.Remote != defaults.LocalRemote {
		t.Errorf("Unexpected remote for default [local] context")
	}

	if config.Context.Endpoint != "http://localhost:3002/" {
		t.Errorf("Unexpected endpoint for default [local] context")
	}

	if config.Context.Username != "no-reply@wedeploy.com" {
		t.Errorf("Unexpected username for default [local] context")
	}

	if config.Context.Password != "cli-tool-password" {
		t.Errorf("Unexpected password for default [local] context")
	}

	if config.Context.Token != "" {
		t.Errorf("Unexpected token for default [local] context")
	}

	Teardown()
}

func TestConfigMockLocalContext(t *testing.T) {
	Setup()
	SetupLocalContext()

	if config.Context.Remote != "" {
		t.Errorf("Unexpected remote for local context")
	}

	if config.Context.Endpoint != "http://localhost:3002/" {
		t.Errorf("Unexpected endpoint for local context")
	}

	if config.Context.Username != "no-reply@wedeploy.com" {
		t.Errorf("Unexpected username for local context")
	}

	if config.Context.Password != "cli-tool-password" {
		t.Errorf("Unexpected password for local context")
	}

	if config.Context.Token != "" {
		t.Errorf("Unexpected token for local context")
	}

	Teardown()
}

func TestConfigMockRemoteContext(t *testing.T) {
	Setup()
	SetupRemoteContext()

	if config.Context.Remote != "foo" {
		t.Errorf("Unexpected remote for remote context")
	}

	if config.Context.Endpoint != "http://www.example.com/" {
		t.Errorf("Unexpected endpoint for remote context")
	}

	if config.Context.Token != "mock_token" {
		t.Errorf("Unexpected token for remote context")
	}

	Teardown()
}
