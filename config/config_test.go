package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wedeploy/cli/remotes"
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
	setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/home"))

	if Global != nil {
		t.Errorf("Expected config.Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected config.Context to be null")
	}

	if err := Setup(); err != nil {
		panic(err)
	}

	if Global.Username != "admin" {
		t.Errorf("Wrong username")
	}

	if Global.Password != "safe" {
		t.Errorf("Wrong password")
	}

	if Global.Local != true {
		t.Errorf("Wrong local value")
	}

	if Global.NotifyUpdates != true {
		t.Errorf("Wrong NotifyUpdate value")
	}

	if Global.ReleaseChannel != "stable" {
		t.Errorf("Wrong ReleaseChannel value")
	}

	if Global.LocalEndpoint != "http://localhost:8080/" {
		t.Errorf("Wrong LocalEndpoint value")
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

	unsetenv("WEDEPLOY_CUSTOM_HOME")
	Teardown()

	if Global != nil {
		t.Errorf("Expected config.Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected config.Context to be null")
	}
}

func TestSetupAndTeardownProject(t *testing.T) {
	setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/home"))
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/non-container")); err != nil {
		t.Error(err)
	}

	if err := Setup(); err != nil {
		panic(err)
	}

	if Context.Scope != "project" {
		t.Errorf("Expected scope to be project, got %v instead", Context.Scope)
	}

	if Context.ProjectRoot != filepath.Join(workingDir, "mocks/project") {
		t.Errorf("Context.ProjectRoot does not match with expected value")
	}

	if Context.ContainerRoot != "" {
		t.Errorf("Expected Context.ContainerRoot to be empty")
	}

	if err := os.Chdir(workingDir); err != nil {
		panic(err)
	}

	unsetenv("WEDEPLOY_CUSTOM_HOME")
	Teardown()
}

func TestSetupAndTeardownProjectAndContainer(t *testing.T) {
	setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/home"))
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/container/inside")); err != nil {
		t.Error(err)
	}

	if err := Setup(); err != nil {
		panic(err)
	}

	if Context.Scope != "container" {
		t.Errorf("Expected scope to be container, got %v instead", Context.Scope)
	}

	if Context.ProjectRoot != filepath.Join(workingDir, "mocks/project") {
		t.Errorf("Context.ProjectRoot does not match with expected value")
	}

	if Context.ContainerRoot != filepath.Join(workingDir, "mocks/project/container") {
		t.Errorf("Expected Context.ContainerRoot to be empty")
	}

	if err := os.Chdir(workingDir); err != nil {
		panic(err)
	}

	unsetenv("WEDEPLOY_CUSTOM_HOME")
	Teardown()
}

func TestSave(t *testing.T) {
	setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/home"))
	if err := Setup(); err != nil {
		panic(err)
	}

	var tmp, err = ioutil.TempFile(os.TempDir(), "we")

	if err != nil {
		panic(err)
	}

	// save in a different location
	Global.Path = tmp.Name()

	Global.Username = "other"

	if err := Global.Save(); err != nil {
		panic(err)
	}

	var got = tdata.FromFile(Global.Path)
	var want = tdata.FromFile("./mocks/we-reference.ini")

	if got != want {
		t.Errorf("Wanted created configuration to match we-reference.ini")
	}

	if err = tmp.Close(); err != nil {
		panic(err)
	}

	if err = os.Remove(tmp.Name()); err != nil {
		panic(err)
	}

	unsetenv("WEDEPLOY_CUSTOM_HOME")
	Teardown()
}

func TestSaveAfterCreation(t *testing.T) {
	setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/homeless"))
	if err := Setup(); err != nil {
		panic(err)
	}

	var tmp, err = ioutil.TempFile(os.TempDir(), "we")

	if err != nil {
		panic(err)
	}

	if Global.Remotes == nil {
		t.Error("Remotes should be initialized, not nil")
	}

	// save in a different location
	Global.Path = tmp.Name()

	Global.Username = "other"

	if err := Global.Save(); err != nil {
		panic(err)
	}

	var got = tdata.FromFile(Global.Path)
	var want = tdata.FromFile("./mocks/we-reference-homeless.ini")

	if got != want {
		t.Errorf("Wanted created configuration to match we-reference-homeless.ini")
	}

	if err = tmp.Close(); err != nil {
		panic(err)
	}

	if err = os.Remove(tmp.Name()); err != nil {
		panic(err)
	}

	unsetenv("WEDEPLOY_CUSTOM_HOME")
	Teardown()
}

func TestRemotes(t *testing.T) {
	setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/remotes"))

	if Global != nil {
		t.Errorf("Expected config.Global to be null")
	}

	if err := Setup(); err != nil {
		panic(err)
	}

	var tmp, err = ioutil.TempFile(os.TempDir(), "we")

	if err != nil {
		panic(err)
	}

	Global.Username = "fool"
	Global.Remotes.Set("staging", "https://staging.example.net/")
	Global.Remotes.Set("beta", "https://beta.example.com/", "remote for beta testing")
	Global.Remotes.Set("new", "http://foo/")
	Global.Remotes.Del("temporary")
	Global.Remotes.Set("remain", "", "commented vars remains even when empty")
	Global.Remotes.Set("dontremain", "")
	Global.Remotes.Del("dontremain2")

	// save in a different location
	Global.Path = tmp.Name()

	if err := Global.Save(); err != nil {
		panic(err)
	}

	var got = tdata.FromFile(Global.Path)

	var want = tdata.FromFile("./mocks/we-reference-remotes.ini")

	if got != want {
		t.Errorf("Wanted created configuration to match we-reference-remotes.ini")
	}

	if err = tmp.Close(); err != nil {
		panic(err)
	}

	if err = os.Remove(tmp.Name()); err != nil {
		panic(err)
	}

	if Global.Username != "fool" {
		t.Errorf("Wrong username")
	}

	unsetenv("WEDEPLOY_CUSTOM_HOME")
	Teardown()

	if Global != nil {
		t.Errorf("Expected config.Global to be null")
	}
}

func TestRemotesListAndGet(t *testing.T) {
	setenv("WEDEPLOY_CUSTOM_HOME", abs("./mocks/remotes"))

	if Global != nil {
		t.Errorf("Expected config.Global to be null")
	}

	if err := Setup(); err != nil {
		panic(err)
	}

	var wantOriginalRemotes = remotes.List{
		"alternative": remotes.Entry{
			URL: "http://example.net/",
		},
		"staging": remotes.Entry{
			URL: "http://staging.example.net/",
		},
		"beta": remotes.Entry{
			URL:        "http://beta.example.com/",
			URLComment: "my beta comment",
		},
		"remain": remotes.Entry{
			URL:     "http://localhost/",
			Comment: "commented vars remains even when empty",
		},
		"dontremain": remotes.Entry{
			URL: "http://localhost/",
		},
		"dontremain2": remotes.Entry{
			URL: "http://localhost/",
		},
	}

	if !reflect.DeepEqual(wantOriginalRemotes, Global.Remotes) {
		t.Errorf("Remotes does not match expected value")
	}

	var wantList = []string{
		"alternative",
		"beta",
		"dontremain",
		"dontremain2",
		"remain",
		"staging",
	}

	var names = Global.Remotes.Keys()

	if !reflect.DeepEqual(names, wantList) {
		t.Errorf("Wanted %v, got %v instead", wantList, names)
	}

	var wantRemain = remotes.Entry{
		URL:     "http://localhost/",
		Comment: "commented vars remains even when empty",
	}

	var gotRemain, gotRemainOK = Global.Remotes["remain"]

	if gotRemain != wantRemain {
		t.Errorf("Wanted %v, got %v instead", wantRemain, gotRemain)
	}

	if !gotRemainOK {
		t.Errorf("Wanted gotRemainOK to be true")
	}

	unsetenv("WEDEPLOY_CUSTOM_HOME")
	Teardown()

	if Global != nil {
		t.Errorf("Expected config.Global to be null")
	}
}

func abs(path string) string {
	var abs, err = filepath.Abs(path)

	if err != nil {
		panic(err)
	}

	return abs
}

func setenv(key, value string) {
	var err = os.Setenv(key, value)

	if err != nil {
		panic(err)
	}
}

func unsetenv(key string) {
	var err = os.Unsetenv(key)

	if err != nil {
		panic(err)
	}
}
