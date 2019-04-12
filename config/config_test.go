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
	wectx, err := Setup("./mocks/invalid/.liferaycli")

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
		wantRemote               = "liferay"
		wantInfrastructureDomain = "us-west-1.liferay.cloud"
	)

	var params = conf.GetParams()

	if len(params.Remotes.Keys()) != 1 {
		t.Errorf("Expected to have one remote, got %v", params.Remotes)
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
	wectx, err := Setup("./mocks/home/.liferaycli")

	if err != nil {
		panic(err)
	}

	conf := wectx.Config()
	params := conf.GetParams()

	if err := wectx.SetEndpoint(defaults.CloudRemote); err != nil {
		panic(err)
	}

	if conf == nil {
		t.Error("Expected config to be mocked")
	}

	var (
		wantUsername       = "foo@example.com"
		wantToken          = "mock_token"
		wantRemote         = "liferay"
		wantInfrastructure = "us-west-1.liferay.cloud"
	)

	if len(params.Remotes.Keys()) != 2 {
		t.Errorf("Expected to have 2 remotes, got %v", params.Remotes)
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
	wectx, err := Setup("./mocks/home/.liferaycli")

	if err != nil {
		panic(err)
	}

	conf := wectx.Config()
	params := conf.GetParams()

	if err := wectx.SetEndpoint(defaults.CloudRemote); err != nil {
		panic(err)
	}

	if conf == nil {
		t.Error("Expected config to be mocked")
	}

	var (
		wantUsername       = "foo@example.com"
		wantToken          = "mock_token"
		wantRemote         = "liferay"
		wantInfrastructure = "us-west-1.liferay.cloud"
	)

	if len(params.Remotes.Keys()) != 2 {
		t.Errorf("Expected to have 2 remotes, got %v", params.Remotes)
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

	if params.NotifyUpdates {
		t.Errorf("Wrong NotifyUpdate value")
	}

	if params.ReleaseChannel != "stable" {
		t.Errorf("Wrong ReleaseChannel value")
	}
}

func TestSetupAndTeardownProject(t *testing.T) {
	var workingDir, _ = os.Getwd()

	if err := os.Chdir(filepath.Join(workingDir, "mocks/project/non-service")); err != nil {
		t.Error(err)
	}

	if _, err := Setup("../../home/.liferaycli"); err != nil {
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

	_, err := Setup("../../../home/.liferaycli")

	if err != nil {
		panic(err)
	}

	if err := os.Chdir(workingDir); err != nil {
		panic(err)
	}
}

func TestSave(t *testing.T) {
	wectx, err := Setup("mocks/home/.liferaycli")

	if err != nil {
		panic(err)
	}

	conf := wectx.Config()

	tmp, err := ioutil.TempFile(os.TempDir(), "liferaycli")

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
		`; Configuration file for Liferay CLI
; https://www.liferay.com/`,
		`default_remote                   = liferay`,
		`local_http_port                  = 80`,
		`local_https_port                 = 443`,
		`disable_autocomplete_autoinstall = true`,
		`disable_colors                   = false`,
		`notify_updates                   = false`,
		`release_channel                  = stable`,
		`enable_analytics                 = false`,
		`[remote "liferay"]
    ; Default cloud remote
    infrastructure = us-west-1.liferay.cloud
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
	wectx, err := Setup("./mocks/remotes/.liferaycli")

	if err != nil {
		panic(err)
	}

	conf := wectx.Config()
	params := conf.GetParams()

	tmp, err := ioutil.TempFile(os.TempDir(), "liferaycli")

	if err != nil {
		panic(err)
	}

	params.Remotes.Set("staging", remotes.Entry{
		Infrastructure: "https://staging.example.net/",
	})

	params.Remotes.Set("beta", remotes.Entry{
		Infrastructure: "https://beta.example.com/",
		Comment:        "remote for beta testing",
	})

	params.Remotes.Set("new", remotes.Entry{
		Infrastructure: "http://foo/",
	})

	params.Remotes.Del("temporary")

	params.Remotes.Set("remain", remotes.Entry{
		Comment: "commented vars remains even when empty",
	})

	params.Remotes.Set("dontremain", remotes.Entry{})

	params.Remotes.Del("dontremain2")

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
[remote "liferay"]
    ; Default cloud remote
    infrastructure = us-west-1.liferay.cloud
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
	wectx, err := Setup("./mocks/remotes/.liferaycli")

	if err != nil {
		panic(err)
	}

	conf := wectx.Config()
	params := conf.GetParams()

	var wantOriginalRemotes = remotes.List{}

	wantOriginalRemotes.Set("liferay", remotes.Entry{
		Infrastructure:        "us-west-1.liferay.cloud",
		InfrastructureComment: "Default cloud remote",
	})

	wantOriginalRemotes.Set("alternative", remotes.Entry{
		Infrastructure: "http://example.net/",
	})

	wantOriginalRemotes.Set("staging", remotes.Entry{
		Infrastructure: "http://staging.example.net/",
	})

	wantOriginalRemotes.Set("beta", remotes.Entry{
		Infrastructure:        "http://beta.example.com/",
		InfrastructureComment: "; my beta comment",
	})

	wantOriginalRemotes.Set("remain", remotes.Entry{
		Infrastructure: "http://localhost/",
		Comment:        "; commented vars remains even when empty",
	})

	wantOriginalRemotes.Set("dontremain", remotes.Entry{
		Infrastructure: "http://localhost/",
		Comment:        "; commented vars remains even when empty",
	})

	wantOriginalRemotes.Set("dontremain2", remotes.Entry{
		Infrastructure: "http://localhost/",
	})

	worKeys := wantOriginalRemotes.Keys()
	prKeys := params.Remotes.Keys()

	if len(worKeys) != len(prKeys) {
		t.Errorf("Number of remotes doesn't match: wanted %v, got %v instead", len(worKeys), len(prKeys))
	}

	for _, k := range wantOriginalRemotes.Keys() {
		want := wantOriginalRemotes.Get(k)
		got := params.Remotes.Get(k)

		if want != got {
			t.Errorf("Expected remote doesn't match for %v: %+v instead of %+v", k, got, want)
		}
	}

	var wantList = []string{
		"alternative",
		"beta",
		"dontremain",
		"dontremain2",
		"liferay",
		"remain",
		"staging",
	}

	var names = params.Remotes.Keys()

	if !reflect.DeepEqual(names, wantList) {
		t.Errorf("Wanted %v, got %v instead", wantList, names)
	}

	var wantRemain = remotes.Entry{
		Infrastructure: "http://localhost/",
		Comment:        "; commented vars remains even when empty",
	}

	gotRemain := params.Remotes.Get("remain")

	if gotRemain != wantRemain {
		t.Errorf("Wanted %v, got %v instead", wantRemain, gotRemain)
	}
}
