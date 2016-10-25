package flagsfromhost

import (
	"reflect"
	"testing"

	"github.com/wedeploy/cli/remotes"
)

var defaultRemotesList = remotesList

func TestParseNoRemoteList(t *testing.T) {
	defer resetDefaults()
	parsed, err := ParseRemoteAddress("foo.bar.wedeploy.me")

	if parsed != "" {
		t.Errorf("Expected remote to be empty, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorLoadingRemoteList:
	default:
		t.Errorf("Expected error to be due to loading remote list, got %v instead", err)
	}

	if err.Error() != "Error loading remotes list" {
		t.Errorf("Wrong error message for parsing when no remote list is loaded")
	}
}

func TestParseRemoteAddressNotFoundForAddress(t *testing.T) {
	defer resetDefaults()
	remotesList = &remotes.List{}
	parsed, err := ParseRemoteAddress("foo.bar.wedeploy.me")

	if parsed != "" {
		t.Errorf("Expected remote to be empty, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorFoundNoRemote:
	default:
		t.Errorf("Expected error to be due to remote not found for address, got %v instead", err)
	}

	if err.Error() != "Found no remote for address foo.bar.wedeploy.me" {
		t.Errorf("Wrong error message for parsing when no remote list is loaded")
	}
}

func TestParseRemoteAddressMultipleRemote(t *testing.T) {
	defer resetDefaults()
	remotesList = &remotes.List{
		"foo": remotes.Entry{
			URL: "foo.bar.wedeploy.me",
		},
		"bar": remotes.Entry{
			URL: "foo.bar.wedeploy.me",
		},
	}

	parsed, err := ParseRemoteAddress("foo.bar.wedeploy.me")

	if parsed != "" {
		t.Errorf("Expected remote to be empty, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorFoundMultipleRemote:
	default:
		t.Errorf("Expected error to be due to multiple remotes found, got %v instead", err)
	}

	if err.Error() != "Found multiple remotes for address foo.bar.wedeploy.me: bar, foo" {
		t.Errorf("Wrong error message for parsing when no remote list is loaded")
	}
}

func TestParseLocalRemoteAddress(t *testing.T) {
	defer resetDefaults()
	remotesList = &remotes.List{
		"foo": remotes.Entry{
			URL: "foo.bar.wedeploy.me",
		},
	}

	if r, err := ParseRemoteAddress(""); r != "" || err != nil {
		t.Errorf("Expected parsed remote to be (empty, nil), got (%v, %v) instead", r, err)
	}

	if r, err := ParseRemoteAddress("wedeploy.me"); r != "" || err != nil {
		t.Errorf("Expected parsed remote to be (empty, nil), got (%v, %v) instead", r, err)
	}
}

func TestParseRemoteAddresses(t *testing.T) {
	defer resetDefaults()
	remotesList = &remotes.List{
		"cloud": remotes.Entry{
			URL: "wedeploy.io",
		},
		"demo": remotes.Entry{
			URL: "demo.wedeploy.com",
		},
	}

	if r, err := ParseRemoteAddress("wedeploy.io"); r != "cloud" || err != nil {
		t.Errorf("Expected parsed remote to be (cloud, nil), got (%v, %v) instead", r, err)
	}
}

type parseByFlagsMockStruct struct {
	project   string
	container string
	remote    string
}

var parseByFlagsMocks = []parseByFlagsMockStruct{
	parseByFlagsMockStruct{
		"cinema",
		"projector",
		"hollywood",
	},
	parseByFlagsMockStruct{
		"cinema",
		"",
		"",
	},
	parseByFlagsMockStruct{
		"cinema",
		"projector",
		"",
	},
	parseByFlagsMockStruct{
		"cinema",
		"",
		"hollywood",
	},
}

func TestParseByFlagsOnly(t *testing.T) {
	defer resetDefaults()
	remotesList = &remotes.List{
		"hollywood": remotes.Entry{
			URL: "example.com",
		},
	}

	for _, k := range parseByFlagsMocks {
		var parsed, err = Parse("", k.project, k.container, k.remote)

		if err != nil {
			t.Errorf("Expected no error on parsing, got %v instead", err)
		}

		if parsed.Project() != k.project ||
			parsed.Container() != k.container ||
			parsed.Remote() != k.remote {
			t.Errorf("Expected values doesn't match on parsed object: %+v", parsed)
		}

		if parsed.IsRemoteFromHost() != false {
			t.Errorf("Expected remote to not be parsed from host")
		}
	}
}

func TestContainerWithMissingProject(t *testing.T) {
	defer resetDefaults()
	remotesList = &remotes.List{
		"hollywood": remotes.Entry{
			URL: "example.com",
		},
	}

	var parsed, err = Parse("", "", "foo", "hollywood")

	var expected = &FlagsFromHost{
		project:   "",
		container: "foo",
		remote:    "hollywood",
	}

	if !reflect.DeepEqual(parsed, expected) {
		t.Errorf("Expected parsed values to be %+v, got %+v instead", parsed, expected)
	}

	switch err.(type) {
	case ErrorContainerWithNoProject:
	default:
		t.Errorf("Expected error type ErrorContainerWithNoProject, got error %v instead", err)
	}

	if err == nil || err.Error() != "Incompatible use: --container requires --project" {
		t.Errorf("Expected no error on parsing, got %v instead", err)
	}
}

func TestParseErrorMultiModeProject(t *testing.T) {
	defer resetDefaults()
	var parsed, err = Parse("foo.bar.wedeploy.com", "a", "", "")

	if parsed != nil {
		t.Errorf("Expected parsed to be nil, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorMultiMode:
	default:
		t.Errorf("Expected err to be of type ErrorMultiMode, got %v instead", err)
	}
}

func TestParseErrorMultiModeContainer(t *testing.T) {
	defer resetDefaults()
	var parsed, err = Parse("foo.bar.wedeploy.com", "", "b", "")

	if parsed != nil {
		t.Errorf("Expected parsed to be nil, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorMultiMode:
	default:
		t.Errorf("Expected err to be of type ErrorMultiMode, got %v instead", err)
	}

	if err.Error() != "Incompatible use: --project and --container are not allowed with host format" {
		t.Errorf("Expected incompatible use message, got %v instead", err)
	}
}

func TestParseHostWithErrorRemoteFlagAndHost(t *testing.T) {
	defer resetDefaults()
	remotesList = &remotes.List{
		"xyz": remotes.Entry{
			URL: "wedeploy.com",
		},
	}

	var parsed, err = Parse("foo.bar.wedeploy.com", "", "", "remote-by-flag")

	if parsed != nil {
		t.Errorf("Expected parsed to be nil, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorRemoteFlagAndHost:
	default:
		t.Errorf("Expected err to be of type ErrorRemoteFlagAndHost, got %v instead", err)
	}

	if err.Error() != "Incompatible use: --remote flag can't be used along host format with remote address" {
		t.Errorf("Expected incompatible use message, got %v instead", err)
	}
}

func TestParseNoMatchFromExternalHost(t *testing.T) {
	defer resetDefaults()
	remotesList = &remotes.List{
		"xyz": remotes.Entry{
			URL: "wedeploy.com",
		},
	}

	var parsed, err = Parse("x.example.com", "", "", "")

	if parsed != nil {
		t.Errorf("Expected parsed to be nil, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorFoundNoRemote:
	default:
		t.Errorf("Expected error to be due to remote not found for address, got %v instead", err)
	}

	if err.Error() != "Found no remote for address x.example.com" {
		t.Errorf("Wrong error message for parsing when no remote list is loaded")
	}
}

type parseHostOnlyMockStruct struct {
	host                 string
	project              string
	container            string
	remote               string
	wantProject          string
	wantContainer        string
	wantRemote           string
	wantIsRemoteFromHost bool
	wantErr              error
}

var parseHostOnlyMocks = []parseHostOnlyMockStruct{
	parseHostOnlyMockStruct{
		host:                 "cinema.example.com",
		project:              "",
		container:            "",
		remote:               "",
		wantProject:          "cinema",
		wantContainer:        "",
		wantRemote:           "foo",
		wantIsRemoteFromHost: true,
		wantErr:              nil,
	},
	parseHostOnlyMockStruct{
		host:                 "projector.cinema.example.com",
		project:              "",
		container:            "",
		remote:               "",
		wantProject:          "cinema",
		wantContainer:        "projector",
		wantRemote:           "foo",
		wantIsRemoteFromHost: true,
		wantErr:              nil,
	},
	parseHostOnlyMockStruct{
		host:                 "projector.cinema",
		project:              "",
		container:            "",
		remote:               "foo",
		wantProject:          "cinema",
		wantContainer:        "projector",
		wantRemote:           "foo",
		wantIsRemoteFromHost: false,
		wantErr:              nil,
	},
	parseHostOnlyMockStruct{
		host:                 "cinema",
		project:              "",
		container:            "",
		remote:               "foo",
		wantProject:          "cinema",
		wantContainer:        "",
		wantRemote:           "foo",
		wantIsRemoteFromHost: false,
		wantErr:              nil,
	},
	parseHostOnlyMockStruct{
		host:                 "abc.def.ghi.jkl.mnn.opq.rst.uvw.xyz",
		project:              "",
		container:            "",
		remote:               "",
		wantProject:          "abc",
		wantContainer:        "",
		wantRemote:           "alphabet",
		wantIsRemoteFromHost: true,
		wantErr:              nil,
	},
	parseHostOnlyMockStruct{
		host:                 "abc.11.22.33.44:5555",
		project:              "",
		container:            "",
		remote:               "",
		wantProject:          "abc",
		wantContainer:        "",
		wantRemote:           "ip",
		wantIsRemoteFromHost: true,
		wantErr:              nil,
	},
	parseHostOnlyMockStruct{
		host:                 "def.abc.11.22.33.44:5555",
		project:              "",
		container:            "",
		remote:               "",
		wantProject:          "abc",
		wantContainer:        "def",
		wantRemote:           "ip",
		wantIsRemoteFromHost: true,
		wantErr:              nil,
	},
	parseHostOnlyMockStruct{
		host:                 "",
		project:              "",
		container:            "",
		remote:               "",
		wantProject:          "",
		wantContainer:        "",
		wantRemote:           "",
		wantIsRemoteFromHost: false,
		wantErr:              nil,
	},
	parseHostOnlyMockStruct{
		host:                 "wedeploy.io",
		project:              "",
		container:            "",
		remote:               "",
		wantProject:          "",
		wantContainer:        "",
		wantRemote:           "wedeploy",
		wantIsRemoteFromHost: true,
		wantErr:              nil,
	},
	parseHostOnlyMockStruct{
		host:                 "wedeploy.me",
		project:              "",
		container:            "",
		remote:               "",
		wantProject:          "",
		wantContainer:        "",
		wantRemote:           "",
		wantIsRemoteFromHost: true,
		wantErr:              nil,
	},
	parseHostOnlyMockStruct{
		host:                 "def.ghi.jkl.mnn.opq.rst.uvw.xyz",
		project:              "",
		container:            "",
		remote:               "",
		wantProject:          "",
		wantContainer:        "",
		wantRemote:           "alphabet",
		wantIsRemoteFromHost: true,
		wantErr:              nil,
	},
}

func TestParse(t *testing.T) {
	defer resetDefaults()
	remotesList = &remotes.List{
		"foo": remotes.Entry{
			URL: "example.com",
		},
		"alphabet": remotes.Entry{
			URL: "def.ghi.jkl.mnn.opq.rst.uvw.xyz",
		},
		"ip": remotes.Entry{
			URL: "11.22.33.44:5555",
		},
		"wedeploy": remotes.Entry{
			URL: "wedeploy.io",
		},
	}

	for _, k := range parseHostOnlyMocks {
		var parsed, err = Parse(k.host, k.project, k.container, k.remote)

		if err != k.wantErr {
			t.Errorf("Expected error to be %v on parsing, got %v instead", k.wantErr, err)
		}

		if parsed.Project() != k.wantProject ||
			parsed.Container() != k.wantContainer ||
			parsed.Remote() != k.wantRemote ||
			parsed.IsRemoteFromHost() != k.wantIsRemoteFromHost {
			t.Errorf("Expected values doesn't match on parsed object: %+v", parsed)
		}
	}
}

func TestParseUnknownRemoteFlagOnly(t *testing.T) {
	defer resetDefaults()
	remotesList = &remotes.List{}
	parsed, err := Parse("", "", "", "cloud")

	if parsed != nil {
		t.Errorf("Expected parsed value to be nil, got %v instead", parsed)
	}

	var wantErr = "Remote cloud not found"

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestParseUnknownRemoteFlag(t *testing.T) {
	defer resetDefaults()
	remotesList = &remotes.List{}
	_, err := Parse("project", "", "", "not-found")

	switch err.(type) {
	case ErrorNotFound:
	default:
		t.Errorf(`Expected error "%v" doesn't match expected type`, err)
	}

	var wantErr = "Remote not-found not found"

	if err.Error() != wantErr {
		t.Errorf("Expected error to be %v on parsing, got %v instead", wantErr, err)
	}
}

func TestParseNoMixingProjectAndHost(t *testing.T) {
	defer resetDefaults()
	switch _, err := Parse("foo.wedeploy.me", "foo", "", ""); err.(type) {
	case ErrorMultiMode:
	default:
		t.Errorf(`Expected error "%v" doesn't match expected type`, err)
	}
}

func TestInjectRemotes(t *testing.T) {
	var list = &remotes.List{}
	InjectRemotes(list)

	if remotesList != list {
		t.Errorf("Expected remotes list is not the same as injected one")
	}
}

func resetDefaults() {
	remotesList = defaultRemotesList
}
