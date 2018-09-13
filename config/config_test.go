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

func TestSetupNonExistingConfigFileAndTeardown(t *testing.T) {
	wectx, err := Setup("./mocks/invalid/.we")

	if err != nil {
		panic(err)
	}

	conf := wectx.Config()

	if err := wectx.SetEndpoint(defaults.CloudRemote); err != nil {
		panic(err)
	}

	if conf == nil {
		t.Error("Expected config to be mocked")
	}

	var (
		wantUsername             = ""
		wantToken                = ""
		wantRemote               = "wedeploy"
		wantInfrastructureDomain = "wedeploy.com"
	)

	if len(conf.Remotes) != 1 {
		t.Errorf("Expected to have one remote, got %v", conf.Remotes)
	}

	if wectx.Username() != wantUsername {
		t.Errorf("Wanted username to be %v, got %v instead", wantUsername, wectx.Username())
	}

	if wectx.Token() != wantToken {
		t.Errorf("Wanted token to be %v, got %v instead", wantToken, wectx.Token())
	}

	if wectx.Remote() != wantRemote {
		t.Errorf("Wanted remote to be %v, got %v instead", wantRemote, wectx.Remote())
	}

	if wectx.InfrastructureDomain() != wantInfrastructureDomain {
		t.Errorf("Wanted InfrastructureDomain to be %v, got %v instead", wantInfrastructureDomain, wectx.InfrastructureDomain())
	}
}

func TestSetupDefaultAndTeardown(t *testing.T) {
	wectx, err := Setup("./mocks/home/.we")

	if err != nil {
		panic(err)
	}

	conf := wectx.Config()

	if err := wectx.SetEndpoint(defaults.CloudRemote); err != nil {
		panic(err)
	}

	if conf == nil {
		t.Error("Expected config to be mocked")
	}

	var (
		wantUsername       = "foo@example.com"
		wantToken          = "mock_token"
		wantRemote         = "wedeploy"
		wantInfrastructure = "wedeploy.com"
	)

	if len(conf.Remotes) != 2 {
		t.Errorf("Expected to have 2 remotes, got %v", conf.Remotes)
	}

	if wectx.Username() != wantUsername {
		t.Errorf("Wanted username to be %v, got %v instead", wantUsername, wectx.Username())
	}

	if wectx.Token() != wantToken {
		t.Errorf("Wanted token to be %v, got %v instead", wantToken, wectx.Token())
	}

	if wectx.Remote() != wantRemote {
		t.Errorf("Wanted remote to be %v, got %v instead", wantRemote, wectx.Remote())
	}

	if wectx.InfrastructureDomain() != wantInfrastructure {
		t.Errorf("Wanted remoteAddress to be %v, got %v instead", wantInfrastructure, wectx.InfrastructureDomain())
	}
}

func TestSetupRemoteAndTeardown(t *testing.T) {
	wectx, err := Setup("./mocks/home/.we")

	if err != nil {
		panic(err)
	}

	conf := wectx.Config()

	if err := wectx.SetEndpoint(defaults.CloudRemote); err != nil {
		panic(err)
	}

	if conf == nil {
		t.Error("Expected config to be mocked")
	}

	var (
		wantUsername       = "foo@example.com"
		wantToken          = "mock_token"
		wantRemote         = "wedeploy"
		wantInfrastructure = "wedeploy.com"
	)

	if len(conf.Remotes) != 2 {
		t.Errorf("Expected to have 2 remotes, got %v", conf.Remotes)
	}

	if wectx.Username() != wantUsername {
		t.Errorf("Wanted username to be %v, got %v instead", wantUsername, wectx.Username())
	}

	if wectx.Token() != wantToken {
		t.Errorf("Wanted token to be %v, got %v instead", wantToken, wectx.Token())
	}

	if wectx.Remote() != wantRemote {
		t.Errorf("Wanted remote to be %v, got %v instead", wantRemote, wectx.Remote())
	}

	if wectx.InfrastructureDomain() != wantInfrastructure {
		t.Errorf("Wanted remoteAddress to be %v, got %v instead", wantInfrastructure, wectx.InfrastructureDomain())
	}

	if conf.NotifyUpdates {
		t.Errorf("Wrong NotifyUpdate value")
	}

	if conf.ReleaseChannel != "stable" {
		t.Errorf("Wrong ReleaseChannel value")
	}
}

func TestSetupAndTeardownProject(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/non-service")); err != nil {
		t.Error(err)
	}

	if _, err := Setup("../../home/.we"); err != nil {
		panic(err)
	}

	if err := os.Chdir(workingDir); err != nil {
		panic(err)
	}
}

func TestSetupAndTeardownProjectAndService(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/service/inside")); err != nil {
		t.Error(err)
	}

	_, err := Setup("../../../home/.we")

	if err != nil {
		panic(err)
	}

	if err := os.Chdir(workingDir); err != nil {
		panic(err)
	}
}

func TestSave(t *testing.T) {
	wectx, err := Setup("mocks/home/.we")

	if err != nil {
		panic(err)
	}

	conf := wectx.Config()

	tmp, err := ioutil.TempFile(os.TempDir(), "we")

	if err != nil {
		panic(err)
	}

	// save in a different location
	conf.Path = tmp.Name()

	if err = conf.Save(); err != nil {
		panic(err)
	}

	var got = tdata.FromFile(conf.Path)
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
    infrastructure = wedeploy.com
    username       = foo@example.com
    token          = mock_token
`,
		`[remote "xyz"]
    infrastructure = wedeploy.xyz
    username       = foobar@example.net
    token          = 123`,
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
}

func TestRemotes(t *testing.T) {
	wectx, err := Setup("./mocks/remotes/.we")

	if err != nil {
		panic(err)
	}

	conf := wectx.Config()

	tmp, err := ioutil.TempFile(os.TempDir(), "we")

	if err != nil {
		panic(err)
	}

	conf.Remotes.Set("staging", remotes.Entry{
		Infrastructure: "https://staging.example.net/",
	})

	conf.Remotes.Set("beta", remotes.Entry{
		Infrastructure: "https://beta.example.com/",
		Comment:        "remote for beta testing",
	})

	conf.Remotes.Set("new", remotes.Entry{
		Infrastructure: "http://foo/",
	})

	conf.Remotes.Del("temporary")

	conf.Remotes.Set("remain", remotes.Entry{
		Comment: "commented vars remains even when empty",
	})

	conf.Remotes.Set("dontremain", remotes.Entry{})

	conf.Remotes.Del("dontremain2")

	// save in a different location
	conf.Path = tmp.Name()

	if err = conf.Save(); err != nil {
		panic(err)
	}

	var got = tdata.FromFile(conf.Path)
	var want = []string{`
[remote "alternative"]
    infrastructure = http://example.net/
`,
		`
[remote "staging"]
    infrastructure = https://staging.example.net/
`,
		`
; remote for beta testing
[remote "beta"]
    infrastructure = https://beta.example.com/
`,
		`
; commented vars remains even when empty
[remote "remain"]
`,
		`
[remote "wedeploy"]
    ; Default cloud remote
    infrastructure = wedeploy.com
`,
		`
[remote "new"]
    infrastructure = http://foo/
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
}

func TestRemotesListAndGet(t *testing.T) {
	wectx, err := Setup("./mocks/remotes/.we")

	if err != nil {
		panic(err)
	}

	conf := wectx.Config()

	var wantOriginalRemotes = remotes.List{
		"wedeploy": remotes.Entry{
			Infrastructure:        "wedeploy.com",
			InfrastructureComment: "Default cloud remote",
		},
		"alternative": remotes.Entry{
			Infrastructure: "http://example.net/",
		},
		"staging": remotes.Entry{
			Infrastructure: "http://staging.example.net/",
		},
		"beta": remotes.Entry{
			Infrastructure:        "http://beta.example.com/",
			InfrastructureComment: "; my beta comment",
		},
		"remain": remotes.Entry{
			Infrastructure: "http://localhost/",
			Comment:        "; commented vars remains even when empty",
		},
		"dontremain": remotes.Entry{
			Infrastructure: "http://localhost/",
			Comment:        "; commented vars remains even when empty",
		},
		"dontremain2": remotes.Entry{
			Infrastructure: "http://localhost/",
		},
	}

	if len(wantOriginalRemotes) != len(conf.Remotes) {
		t.Errorf("Number of remotes doesn't match: wanted %v, got %v instead", len(wantOriginalRemotes), len(conf.Remotes))
	}

	for k := range wantOriginalRemotes {
		if wantOriginalRemotes[k] != conf.Remotes[k] {
			t.Errorf("Expected remote doesn't match for %v: %+v instead of %+v", k, conf.Remotes[k], wantOriginalRemotes[k])
		}
	}

	var wantList = []string{
		"alternative",
		"beta",
		"dontremain",
		"dontremain2",
		"remain",
		"staging",
		"wedeploy",
	}

	var names = conf.Remotes.Keys()

	if !reflect.DeepEqual(names, wantList) {
		t.Errorf("Wanted %v, got %v instead", wantList, names)
	}

	var wantRemain = remotes.Entry{
		Infrastructure: "http://localhost/",
		Comment:        "; commented vars remains even when empty",
	}

	var gotRemain, gotRemainOK = conf.Remotes["remain"]

	if gotRemain != wantRemain {
		t.Errorf("Wanted %v, got %v instead", wantRemain, gotRemain)
	}

	if !gotRemainOK {
		t.Errorf("Wanted gotRemainOK to be true")
	}
}
