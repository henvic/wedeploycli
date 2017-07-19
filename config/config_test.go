package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/remotes"
	"github.com/wedeploy/cli/tdata"
)

func TestUnset(t *testing.T) {
	if Global != nil {
		t.Errorf("Expected Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected Context to be null")
	}
}

func TestSetupNonExistingConfigFileAndTeardown(t *testing.T) {
	if Global != nil {
		t.Errorf("Expected Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected Context to be null")
	}

	if err := Setup("./mocks/invalid/.we"); err != nil {
		panic(err)
	}

	if err := SetEndpointContext(defaults.CloudRemote); err != nil {
		panic(err)
	}

	if Global == nil {
		t.Error("Expected global config to be mocked")
	}

	var (
		wantUsername      = ""
		wantPassword      = ""
		wantToken         = ""
		wantRemote        = "wedeploy"
		wantRemoteAddress = "wedeploy.io"
	)

	if len(Global.Remotes) != 2 {
		t.Errorf("Expected to have 2 remotes, got %v", Global.Remotes)
	}

	if Context.Username != wantUsername {
		t.Errorf("Wanted username to be %v, got %v instead", wantUsername, Context.Username)
	}

	if Context.Password != wantPassword {
		t.Errorf("Wanted password to be %v, got %v instead", wantPassword, Context.Password)
	}

	if Context.Token != wantToken {
		t.Errorf("Wanted token to be %v, got %v instead", wantToken, Context.Token)
	}

	if Context.Remote != wantRemote {
		t.Errorf("Wanted remote to be %v, got %v instead", wantRemote, Context.Remote)
	}

	if Context.RemoteAddress != wantRemoteAddress {
		t.Errorf("Wanted remoteAddress to be %v, got %v instead", wantRemoteAddress, Context.RemoteAddress)
	}

	Teardown()

	if Global != nil {
		t.Errorf("Expected Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected Context to be null")
	}
}

func TestSetupLocalAndTeardown(t *testing.T) {
	if Global != nil {
		t.Errorf("Expected Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected Context to be null")
	}

	if err := Setup("./mocks/home/.we"); err != nil {
		panic(err)
	}

	if err := SetEndpointContext(defaults.LocalRemote); err != nil {
		panic(err)
	}

	if Global == nil {
		t.Error("Expected global config to be mocked")
	}

	var (
		wantUsername      = "foo@example.com"
		wantPassword      = ""
		wantToken         = "mock_token"
		wantRemote        = "local"
		wantRemoteAddress = "wedeploy.me"
	)

	if len(Global.Remotes) != 3 {
		t.Errorf("Expected to have 3 remotes, got %v", Global.Remotes)
	}

	if Context.Username != wantUsername {
		t.Errorf("Wanted username to be %v, got %v instead", wantUsername, Context.Username)
	}

	if Context.Password != wantPassword {
		t.Errorf("Wanted password to be %v, got %v instead", wantPassword, Context.Password)
	}

	if Context.Token != wantToken {
		t.Errorf("Wanted token to be %v, got %v instead", wantToken, Context.Token)
	}

	if Context.Remote != wantRemote {
		t.Errorf("Wanted remote to be %v, got %v instead", wantRemote, Context.Remote)
	}

	if Context.RemoteAddress != wantRemoteAddress {
		t.Errorf("Wanted remoteAddress to be %v, got %v instead", wantRemoteAddress, Context.RemoteAddress)
	}

	Teardown()

	if Global != nil {
		t.Errorf("Expected Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected Context to be null")
	}
}

func TestSetupRemoteAndTeardown(t *testing.T) {
	if Global != nil {
		t.Errorf("Expected Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected config.Context to be null")
	}

	if err := Setup("./mocks/home/.we"); err != nil {
		panic(err)
	}

	if err := SetEndpointContext(defaults.CloudRemote); err != nil {
		panic(err)
	}

	if Global == nil {
		t.Error("Expected global config to be mocked")
	}

	var (
		wantUsername      = "foo@example.com"
		wantPassword      = "bar"
		wantToken         = ""
		wantRemote        = "wedeploy"
		wantRemoteAddress = "wedeploy.io"
	)

	if len(Global.Remotes) != 3 {
		t.Errorf("Expected to have 3 remotes, got %v", Global.Remotes)
	}

	if Context.Username != wantUsername {
		t.Errorf("Wanted username to be %v, got %v instead", wantUsername, Context.Username)
	}

	if Context.Password != wantPassword {
		t.Errorf("Wanted password to be %v, got %v instead", wantPassword, Context.Password)
	}

	if Context.Token != wantToken {
		t.Errorf("Wanted token to be %v, got %v instead", wantToken, Context.Token)
	}

	if Context.Remote != wantRemote {
		t.Errorf("Wanted remote to be %v, got %v instead", wantRemote, Context.Remote)
	}

	if Context.RemoteAddress != wantRemoteAddress {
		t.Errorf("Wanted remoteAddress to be %v, got %v instead", wantRemoteAddress, Context.RemoteAddress)
	}

	if Global.NotifyUpdates {
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

	if Context.ServiceRoot != "" {
		t.Errorf("Expected Context.ServiceRoot to be empty")
	}

	Teardown()

	if Global != nil {
		t.Errorf("Expected Global to be null")
	}

	if Context != nil {
		t.Errorf("Expected config.Context to be null")
	}
}

func TestSetupAndTeardownProject(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/non-service")); err != nil {
		t.Error(err)
	}

	if err := Setup("../../home/.we"); err != nil {
		panic(err)
	}

	if Context.Scope != "project" {
		t.Errorf("Expected scope to be project, got %v instead", Context.Scope)
	}

	if Context.ProjectRoot != filepath.Join(workingDir, "mocks/project") {
		t.Errorf("Context.ProjectRoot does not match with expected value")
	}

	if Context.ServiceRoot != "" {
		t.Errorf("Expected Context.ServiceRoot to be empty")
	}

	if err := os.Chdir(workingDir); err != nil {
		panic(err)
	}

	Teardown()
}

func TestSetupAndTeardownProjectAndService(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/service/inside")); err != nil {
		t.Error(err)
	}

	if err := Setup("../../../home/.we"); err != nil {
		panic(err)
	}

	if Context.Scope != "service" {
		t.Errorf("Expected scope to be service, got %v instead", Context.Scope)
	}

	if Context.ProjectRoot != filepath.Join(workingDir, "mocks/project") {
		t.Errorf("Context.ProjectRoot does not match with expected value")
	}

	if Context.ServiceRoot != filepath.Join(workingDir, "mocks/project/service") {
		t.Errorf("Expected Context.ServiceRoot to be empty")
	}

	if err := os.Chdir(workingDir); err != nil {
		panic(err)
	}

	Teardown()
}

func TestSave(t *testing.T) {
	if err := Setup("mocks/home/.we"); err != nil {
		panic(err)
	}

	var tmp, err = ioutil.TempFile(os.TempDir(), "we")

	if err != nil {
		panic(err)
	}

	// save in a different location
	Global.Path = tmp.Name()

	if err := Global.Save(); err != nil {
		panic(err)
	}

	var got = tdata.FromFile(Global.Path)
	var want = []string{
		`; Configuration file for WeDeploy CLI
; https://wedeploy.com`,
		`default_remote                   = wedeploy`,
		`local_http_port                  = 80`,
		`local_https_port                 = 443`,
		`disable_autocomplete_autoinstall = true`,
		`disable_colors                   = false`,
		`notify_updates                   = false`,
		`release_channel                  = stable`,
		`enable_analytics                 = false`,
		`[remote "wedeploy"]
    ; Default cloud remote
    url      = wedeploy.io
    username = foo@example.com
    password = bar
`,
		`[remote "local"]
    ; Default local remote
    url      = http://wedeploy.me
    username = foo@example.com
    token    = mock_token
`,
		`[remote "xyz"]
    url      = wedeploy.xyz
    username = foobar@example.net
    password = 123`,
	}

	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("Expected string does not exists in generated configuration file: %v", w)
		}
	}

	if err = tmp.Close(); err != nil {
		panic(err)
	}

	if err = os.Remove(tmp.Name()); err != nil {
		panic(err)
	}

	Teardown()
}

func TestRemotes(t *testing.T) {
	if Global != nil {
		t.Errorf("Expected Global to be null")
	}

	if err := Setup("./mocks/remotes/.we"); err != nil {
		panic(err)
	}

	var tmp, err = ioutil.TempFile(os.TempDir(), "we")

	if err != nil {
		panic(err)
	}

	Global.Remotes.Set("staging", remotes.Entry{
		URL: "https://staging.example.net/",
	})

	Global.Remotes.Set("beta", remotes.Entry{
		URL:     "https://beta.example.com/",
		Comment: "remote for beta testing",
	})

	Global.Remotes.Set("new", remotes.Entry{
		URL: "http://foo/",
	})

	Global.Remotes.Del("temporary")

	Global.Remotes.Set("remain", remotes.Entry{
		Comment: "commented vars remains even when empty",
	})

	Global.Remotes.Set("dontremain", remotes.Entry{})

	Global.Remotes.Del("dontremain2")

	// save in a different location
	Global.Path = tmp.Name()

	if err := Global.Save(); err != nil {
		panic(err)
	}

	var got = tdata.FromFile(Global.Path)
	var want = []string{`
[remote "alternative"]
    url = http://example.net/
`,
		`
[remote "staging"]
    url = https://staging.example.net/
`,
		`
; remote for beta testing
[remote "beta"]
    url = https://beta.example.com/
`,
		`
; commented vars remains even when empty
[remote "remain"]
`,
		`
[remote "wedeploy"]
    ; Default cloud remote
    url = wedeploy.io
`,
		`
[remote "new"]
    url = http://foo/
`}

	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("Expected string does not exists in generated configuration file: %v", w)
		}
	}

	if err = tmp.Close(); err != nil {
		panic(err)
	}

	if err = os.Remove(tmp.Name()); err != nil {
		panic(err)
	}

	Teardown()

	if Global != nil {
		t.Errorf("Expected Global to be null")
	}
}

func TestRemotesListAndGet(t *testing.T) {
	if Global != nil {
		t.Errorf("Expected Global to be null")
	}

	if err := Setup("./mocks/remotes/.we"); err != nil {
		panic(err)
	}

	var wantOriginalRemotes = remotes.List{
		"wedeploy": remotes.Entry{
			URL:        "wedeploy.io",
			URLComment: "Default cloud remote",
		},
		"local": remotes.Entry{
			URL:        "http://wedeploy.me",
			URLComment: "Default local remote",
			Username:   "no-reply@wedeploy.com",
			Password:   "cli-tool-password",
		},
		"alternative": remotes.Entry{
			URL: "http://example.net/",
		},
		"staging": remotes.Entry{
			URL: "http://staging.example.net/",
		},
		"beta": remotes.Entry{
			URL:        "http://beta.example.com/",
			URLComment: "; my beta comment",
		},
		"remain": remotes.Entry{
			URL:     "http://localhost/",
			Comment: "; commented vars remains even when empty",
		},
		"dontremain": remotes.Entry{
			URL:     "http://localhost/",
			Comment: "; commented vars remains even when empty",
		},
		"dontremain2": remotes.Entry{
			URL: "http://localhost/",
		},
	}

	if len(wantOriginalRemotes) != len(Global.Remotes) {
		t.Errorf("Number of remotes doesn't match: wanted %v, got %v instead", len(wantOriginalRemotes), len(Global.Remotes))
	}

	for k := range wantOriginalRemotes {
		if wantOriginalRemotes[k] != Global.Remotes[k] {
			t.Errorf("Expected remote doesn't match for %v: %+v instead of %+v", k, Global.Remotes[k], wantOriginalRemotes[k])
		}
	}

	var wantList = []string{
		"alternative",
		"beta",
		"dontremain",
		"dontremain2",
		"local",
		"remain",
		"staging",
		"wedeploy",
	}

	var names = Global.Remotes.Keys()

	if !reflect.DeepEqual(names, wantList) {
		t.Errorf("Wanted %v, got %v instead", wantList, names)
	}

	var wantRemain = remotes.Entry{
		URL:     "http://localhost/",
		Comment: "; commented vars remains even when empty",
	}

	var gotRemain, gotRemainOK = Global.Remotes["remain"]

	if gotRemain != wantRemain {
		t.Errorf("Wanted %v, got %v instead", wantRemain, gotRemain)
	}

	if !gotRemainOK {
		t.Errorf("Wanted gotRemainOK to be true")
	}

	Teardown()

	if Global != nil {
		t.Errorf("Expected Global to be null")
	}
}

func abs(path string) string {
	var abs, err = filepath.Abs(path)

	if err != nil {
		panic(err)
	}

	return abs
}
