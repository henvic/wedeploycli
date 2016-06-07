package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/wedeploy/cli/tdata"
)

func TestUnset(t *testing.T) {
	if Global != nil {
		t.Errorf("Expected config.Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected config.Context to be null")
	}
}

func TestSetupAndTeardown(t *testing.T) {
	os.Setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/home"))

	if Global != nil {
		t.Errorf("Expected config.Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected config.Context to be null")
	}

	Setup()

	if Global.Username != "admin" {
		t.Errorf("Wrong username")
	}

	if Global.Password != "safe" {
		t.Errorf("Wrong password")
	}

	if Global.Local != true {
		t.Errorf("Wrong local value")
	}

	if Global.Endpoint != "http://www.example.com/" {
		t.Errorf("Wrong endpoint")
	}

	if Global.NotifyUpdates != true {
		t.Errorf("Wrong NotifyUpdate value")
	}

	if Global.ReleaseChannel != "stable" {
		t.Errorf("Wrong ReleaseChannel value")
	}

	if Context.Scope != "global" {
		t.Errorf("Exected global scope")
	}

	if Context.ProjectRoot != "" {
		t.Errorf("Expected Context.ProjectRoot to be empty")
	}

	if Context.ContainerRoot != "" {
		t.Errorf("Expected Context.ContainerRoot to be empty")
	}

	os.Unsetenv("WEDEPLOY_CUSTOM_HOME")
	Teardown()

	if Global != nil {
		t.Errorf("Expected config.Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected config.Context to be null")
	}
}

func TestSetupAndTeardownProject(t *testing.T) {
	os.Setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/home"))
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/non-container")); err != nil {
		t.Error(err)
	}

	Setup()

	if Context.Scope != "project" {
		t.Errorf("Expected scope to be project, got %v instead", Context.Scope)
	}

	if Context.ProjectRoot != filepath.Join(workingDir, "mocks/project") {
		t.Errorf("Context.ProjectRoot doesn't match with expected value")
	}

	if Context.ContainerRoot != "" {
		t.Errorf("Expected Context.ContainerRoot to be empty")
	}

	if err := os.Chdir(workingDir); err != nil {
		panic(err)
	}

	os.Unsetenv("WEDEPLOY_CUSTOM_HOME")
	Teardown()
}

func TestSetupAndTeardownProjectAndContainer(t *testing.T) {
	os.Setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/home"))
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/container/inside")); err != nil {
		t.Error(err)
	}

	Setup()

	if Context.Scope != "container" {
		t.Errorf("Expected scope to be container, got %v instead", Context.Scope)
	}

	if Context.ProjectRoot != filepath.Join(workingDir, "mocks/project") {
		t.Errorf("Context.ProjectRoot doesn't match with expected value")
	}

	if Context.ContainerRoot != filepath.Join(workingDir, "mocks/project/container") {
		t.Errorf("Expected Context.ContainerRoot to be empty")
	}

	if err := os.Chdir(workingDir); err != nil {
		panic(err)
	}

	os.Unsetenv("WEDEPLOY_CUSTOM_HOME")
	Teardown()
}

func TestSave(t *testing.T) {
	os.Setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/home"))
	Setup()

	var tmp, err = ioutil.TempFile(os.TempDir(), "we")

	if err != nil {
		panic(err)
	}

	// save in a different location
	Global.Path = tmp.Name()

	Global.Username = "other"
	Global.Save()

	var got = tdata.FromFile(Global.Path)
	var want = tdata.FromFile("./mocks/we-reference.ini")

	if got != want {
		t.Errorf("Wanted created configuration to match we-reference.ini")
	}

	tmp.Close()
}

func TestSaveAfterCreation(t *testing.T) {
	os.Setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/homeless"))
	Setup()

	var tmp, err = ioutil.TempFile(os.TempDir(), "we")

	if err != nil {
		panic(err)
	}

	// save in a different location
	Global.Path = tmp.Name()

	Global.Username = "other"
	Global.Save()

	var got = tdata.FromFile(Global.Path)
	var want = tdata.FromFile("./mocks/we-reference-homeless.ini")

	if got != want {
		t.Errorf("Wanted created configuration to match we-reference-homeless.ini")
	}

	tmp.Close()
}

func abs(path string) string {
	var abs, err = filepath.Abs(path)

	if err != nil {
		panic(err)
	}

	return abs
}
