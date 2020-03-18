package flagsfromhost

import (
	"reflect"
	"testing"

	"github.com/henvic/wedeploycli/remotes"
)

func TestParseNoRemoteList(t *testing.T) {
	var c = CommandFlagFromHost{}
	parsed, err := c.ParseRemoteAddress("foo-bar.wedeploy.me")

	if parsed != "" {
		t.Errorf("Expected remote to be empty, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorLoadingRemoteList:
	default:
		t.Errorf("Expected error to be due to loading remote list, got %v instead", err)
	}

	if err.Error() != "error loading remotes list" {
		t.Errorf("Wrong error message for parsing when no remote list is loaded")
	}
}
func TestParseRemoteAddressNotFoundForAddress(t *testing.T) {
	var c = New(&remotes.List{})

	parsed, err := c.ParseRemoteAddress("foo-bar.wedeploy.me")

	if parsed != "" {
		t.Errorf("Expected remote to be empty, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorFoundNoRemote:
	default:
		t.Errorf("Expected error to be due to remote not found for address, got %v instead", err)
	}

	if err.Error() != "found no remote for address foo-bar.wedeploy.me" {
		t.Errorf("Wrong error message for parsing when no remote list is loaded")
	}
}

func TestParseRemoteAddressMultipleRemote(t *testing.T) {
	var rl = remotes.List{}

	rl.Set("foo", remotes.Entry{
		Service: "foo-bar.wedeploy.me",
	})

	rl.Set("bar", remotes.Entry{
		Service: "foo-bar.wedeploy.me",
	})

	var c = New(&rl)

	parsed, err := c.ParseRemoteAddress("foo-bar.wedeploy.me")

	if parsed != "" {
		t.Errorf("Expected remote to be empty, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorFoundMultipleRemote:
	default:
		t.Errorf("Expected error to be due to multiple remotes found, got %v instead", err)
	}

	if err.Error() != "found multiple remotes for address foo-bar.wedeploy.me: bar, foo" {
		t.Errorf("Wrong error message for parsing when no remote list is loaded")
	}
}

func TestParseMultipleRemoteHostMultipleRemote(t *testing.T) {
	var rl = remotes.List{}

	rl.Set("foo", remotes.Entry{
		Service: "wedeploy.me",
	})

	rl.Set("bar", remotes.Entry{
		Service: "wedeploy.me",
	})

	var c = New(&rl)

	parsed, err := c.Parse(ParseFlags{
		Host: "wedeploy.me",
	})

	if parsed != nil {
		t.Errorf("Expected remote to be nil, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorFoundMultipleRemote:
	default:
		t.Errorf("Expected error to be due to multiple remotes found, got %v instead", err)
	}

	if err != nil && err.Error() != "found multiple remotes for address wedeploy.me: bar, foo" {
		t.Errorf("Wrong error message for parsing when no remote list is loaded, got %v instead", err)
	}
}

func TestParseLocalRemoteAddress(t *testing.T) {
	var rl = remotes.List{}

	rl.Set("foo", remotes.Entry{
		Service: "foo-bar.wedeploy.me",
	})

	var c = New(&rl)

	if r, err := c.ParseRemoteAddress(""); r != "" || err != nil {
		t.Errorf("Expected parsed remote to be (empty, nil), got (%v, %v) instead", r, err)
	}
}

func TestParseRemoteAddresses(t *testing.T) {
	var rl = remotes.List{}

	rl.Set("cloud", remotes.Entry{
		Service: "lfr.cloud",
	})

	rl.Set("demo", remotes.Entry{
		Service: "demo.wedeploy.com",
	})

	var c = New(&rl)

	if r, err := c.ParseRemoteAddress("lfr.cloud"); r != "cloud" || err != nil {
		t.Errorf("Expected parsed remote to be (cloud, nil), got (%v, %v) instead", r, err)
	}
}

var parseByFlagsMocks = []ParseFlags{
	ParseFlags{
		"cinema",
		"projector",
		"hollywood",
		"",
	},
	ParseFlags{
		"cinema",
		"",
		"",
		"",
	},
	ParseFlags{
		"cinema",
		"projector",
		"",
		"",
	},
	ParseFlags{
		"cinema",
		"",
		"hollywood",
		"",
	},
}

func TestParseByFlagsOnly(t *testing.T) {
	var rl = remotes.List{}

	rl.Set("hollywood", remotes.Entry{
		Service: "example.com",
	})

	var c = New(&rl)

	for _, k := range parseByFlagsMocks {
		var parsed, err = c.Parse(k)

		if err != nil {
			t.Errorf("Expected no error on parsing, got %v instead", err)
		}

		if parsed.Project() != k.Project ||
			parsed.Service() != k.Service ||
			parsed.Remote() != k.Remote {
			t.Errorf("Expected values doesn't match on parsed object: %+v", parsed)
		}

		if parsed.IsRemoteFromHost() {
			t.Errorf("Expected remote to not be parsed from host")
		}
	}
}

func TestServiceWithMissingProject(t *testing.T) {
	var rl = remotes.List{}

	rl.Set("hollywood", remotes.Entry{
		Service: "example.com",
	})

	var c = New(&rl)

	var parsed, err = c.Parse(ParseFlags{
		Service: "foo",
		Remote:  "hollywood",
	})

	var expected = &FlagsFromHost{
		project: "",
		service: "foo",
		remote:  "hollywood",
	}

	if !reflect.DeepEqual(parsed, expected) {
		t.Errorf("Expected parsed values to be %+v, got %+v instead", parsed, expected)
	}

	switch err.(type) {
	case ErrorServiceWithNoProject:
	default:
		t.Errorf("Expected error type ErrorServiceWithNoProject, got error %v instead", err)
	}

	if err == nil || err.Error() != "incompatible use: --service requires --project" {
		t.Errorf("Expected no error on parsing, got %v instead", err)
	}
}

func TestParseErrorMultiModeProject(t *testing.T) {
	var c = CommandFlagFromHost{}
	var parsed, err = c.Parse(ParseFlags{
		Host:    "foo-bar.wedeploy.com",
		Project: "a",
	})

	if parsed != nil {
		t.Errorf("Expected parsed to be nil, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorMultiMode:
	default:
		t.Errorf("Expected err to be of type ErrorMultiMode, got %v instead", err)
	}
}

func TestParseErrorMultiModeService(t *testing.T) {
	var c = CommandFlagFromHost{}
	var parsed, err = c.Parse(ParseFlags{
		Host:    "foo-bar.wedeploy.com",
		Service: "b",
	})

	if parsed != nil {
		t.Errorf("Expected parsed to be nil, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorMultiMode:
	default:
		t.Errorf("Expected err to be of type ErrorMultiMode, got %v instead", err)
	}

	if err.Error() != "incompatible use: --project and --service are not allowed with host URL flag" {
		t.Errorf("Expected incompatible use message, got %v instead", err)
	}
}

func TestParseHostWithErrorRemoteFlagAndHost(t *testing.T) {
	var rl = remotes.List{}

	rl.Set("xyz", remotes.Entry{
		Service: "wedeploy.com",
	})

	var c = New(&rl)

	var parsed, err = c.Parse(ParseFlags{
		Host:   "foo-bar.wedeploy.com",
		Remote: "remote",
	})

	if parsed != nil {
		t.Errorf("Expected parsed to be nil, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorRemoteFlagAndHost:
	default:
		t.Errorf("Expected err to be of type ErrorRemoteFlagAndHost, got %v instead", err)
	}

	if err.Error() != "incompatible use: --remote flag cannot be used along host format with remote address" {
		t.Errorf("Expected incompatible use message, got %v instead", err)
	}
}

func TestParseNoMatchFromExternalHost(t *testing.T) {
	var rl = remotes.List{}

	rl.Set("xyz", remotes.Entry{
		Service: "wedeploy.com",
	})

	var c = New(&rl)

	var parsed, err = c.Parse(ParseFlags{
		Host: "x.example.com",
	})

	if parsed != nil {
		t.Errorf("Expected parsed to be nil, got %v instead", parsed)
	}

	switch err.(type) {
	case ErrorFoundNoRemote:
	default:
		t.Errorf("Expected error to be due to remote not found for address, got %v instead", err)
	}

	if err.Error() != "found no remote for address x.example.com" {
		t.Errorf("Wrong error message for parsing when no remote list is loaded")
	}
}

type parseMock struct {
	Flags ParseFlags
	Want  parsed
}

type parsed struct {
	Project          string
	Service          string
	Remote           string
	IsRemoteFromHost bool
	Err              error
}

var parseMocks = []parseMock{
	parseMock{
		Flags: ParseFlags{
			Host:    "EXAMPLE.com",
			Project: "",
			Service: "",
			Remote:  "foo",
		},
		Want: parsed{
			Err: ErrorRemoteFlagAndHost{},
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "example.com",
			Project: "",
			Service: "",
			Remote:  "foo",
		},
		Want: parsed{
			Err: ErrorRemoteFlagAndHost{},
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "CINEMA.EXAMPLE.COM",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "cinema",
			Remote:           "foo",
			IsRemoteFromHost: true,
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "cinema.example.com",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "cinema",
			Remote:           "foo",
			IsRemoteFromHost: true,
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "projector-cinema.example.com",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "cinema",
			Service:          "projector",
			Remote:           "foo",
			IsRemoteFromHost: true,
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "projector-smart-cinema.example.com",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "smart-cinema",
			Service:          "projector",
			Remote:           "foo",
			IsRemoteFromHost: true,
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "projector-laser-3d-cinema.example.com",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "laser-3d-cinema",
			Service:          "projector",
			Remote:           "foo",
			IsRemoteFromHost: true,
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "projector-cinema",
			Project: "",
			Service: "",
			Remote:  "foo",
		},
		Want: parsed{
			Project: "cinema",
			Service: "projector",
			Remote:  "foo",
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "cinema",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Service: "cinema",
			Err:     ErrorServiceWithNoProject{},
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "cinema",
			Project: "",
			Service: "",
			Remote:  "foo",
		},
		Want: parsed{
			Service: "cinema",
			Remote:  "foo",
			Err:     ErrorServiceWithNoProject{},
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "abc.def.ghi.jkl.mnn.opq.rst.uvw.xyz",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "abc",
			Remote:           "alphabet",
			IsRemoteFromHost: true,
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "abc.11.22.33.44:5555",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "abc",
			Remote:           "ip",
			IsRemoteFromHost: true,
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "def-abc.11.22.33.44:5555",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "abc",
			Service:          "def",
			Remote:           "ip",
			IsRemoteFromHost: true,
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "lfr.cloud",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Remote:           "lcp",
			IsRemoteFromHost: true,
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "lfr.cloud",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			IsRemoteFromHost: true,
			Remote:           "lcp",
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "foo.lfr.cloud",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "foo",
			Remote:           "lcp",
			IsRemoteFromHost: true,
		},
	},
	parseMock{
		Flags: ParseFlags{
			Host:    "def.ghi.jkl.mnn.opq.rst.uvw.xyz",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Remote:           "alphabet",
			IsRemoteFromHost: true,
		},
	},
}

func testParse(c CommandFlagFromHost, pm parseMock, t *testing.T) {
	var parsed, err = c.Parse(pm.Flags)

	if err != pm.Want.Err {
		t.Errorf("Expected error to be %v on parsing, got %v instead", pm.Want.Err, err)
	}

	if parsed != nil && (parsed.Project() != pm.Want.Project ||
		parsed.Service() != pm.Want.Service ||
		parsed.Remote() != pm.Want.Remote ||
		parsed.IsRemoteFromHost() != pm.Want.IsRemoteFromHost) {
		t.Errorf("Expected values doesn't match on parsed object: %+v", parsed)
	}
}

func TestParse(t *testing.T) {
	var rl = remotes.List{}

	rl.Set("foo", remotes.Entry{
		Service: "example.com",
	})

	rl.Set("alphabet", remotes.Entry{
		Service: "def.ghi.jkl.mnn.opq.rst.uvw.xyz",
	})

	rl.Set("ip", remotes.Entry{
		Service: "11.22.33.44:5555",
	})

	rl.Set("lcp", remotes.Entry{
		Service: "lfr.cloud",
	})

	var c = New(&rl)

	for _, k := range parseMocks {
		testParse(c, k, t)
	}
}

type parseMockWithDefaultCustomRemote struct {
	Flags ParseFlagsWithDefaultCustomRemote
	Want  parsed
}

var parseMocksWithDefaultCustomRemote = []parseMockWithDefaultCustomRemote{
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:          "",
			Project:       "",
			Service:       "",
			Remote:        "",
			RemoteChanged: true,
		},
		Want: parsed{
			Remote: "lcp",
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:          "example.com",
			Project:       "",
			Service:       "",
			Remote:        "foo",
			RemoteChanged: true,
		},
		Want: parsed{
			Err: ErrorRemoteFlagAndHost{},
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:          "",
			Project:       "CINEMA",
			Service:       "THEATER",
			Remote:        "FOO",
			RemoteChanged: true,
		},
		Want: parsed{
			Project: "cinema",
			Service: "theater",
			Remote:  "foo",
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "cinema.example.com",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "cinema",
			Remote:           "foo",
			IsRemoteFromHost: true,
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "projector-cinema.example.com",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "cinema",
			Service:          "projector",
			Remote:           "foo",
			IsRemoteFromHost: true,
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:          "projector-cinema",
			Project:       "",
			Service:       "",
			Remote:        "foo",
			RemoteChanged: true,
		},
		Want: parsed{
			Project: "cinema",
			Service: "projector",
			Remote:  "foo",
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "cinema",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Service: "cinema",
			Remote:  "lcp",
			Err:     ErrorServiceWithNoProject{},
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:          "cinema",
			Project:       "",
			Service:       "",
			Remote:        "foo",
			RemoteChanged: true,
		},
		Want: parsed{
			Service: "cinema",
			Remote:  "foo",
			Err:     ErrorServiceWithNoProject{},
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "abc.def.ghi.jkl.mnn.opq.rst.uvw.xyz",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "abc",
			Remote:           "alphabet",
			IsRemoteFromHost: true,
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "abc.11.22.33.44:5555",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "abc",
			Remote:           "ip",
			IsRemoteFromHost: true,
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "def-abc.11.22.33.44:5555",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "abc",
			Service:          "def",
			Remote:           "ip",
			IsRemoteFromHost: true,
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Remote: "lcp",
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "lfr.cloud",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Remote:           "lcp",
			IsRemoteFromHost: true,
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "wedeploy.me",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Remote:           "local",
			IsRemoteFromHost: true,
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "project.wedeploy.me",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "project",
			Remote:           "local",
			IsRemoteFromHost: true,
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "service-project.wedeploy.me",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Project:          "project",
			Service:          "service",
			Remote:           "local",
			IsRemoteFromHost: true,
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:          "",
			Project:       "",
			Service:       "",
			Remote:        "local",
			RemoteChanged: true,
		},
		Want: parsed{
			Remote: "local",
		},
	},
	parseMockWithDefaultCustomRemote{
		Flags: ParseFlagsWithDefaultCustomRemote{
			Host:    "def.ghi.jkl.mnn.opq.rst.uvw.xyz",
			Project: "",
			Service: "",
			Remote:  "",
		},
		Want: parsed{
			Remote:           "alphabet",
			IsRemoteFromHost: true,
		},
	},
}

func testParseWithDefaultCustomRemote(c CommandFlagFromHost, pm parseMockWithDefaultCustomRemote, t *testing.T) {
	var parsed, err = c.ParseWithDefaultCustomRemote(pm.Flags, "lcp")

	if err != pm.Want.Err {
		t.Errorf("Expected error to be %v on parsing, got %v instead", pm.Want.Err, err)
	}

	if parsed != nil && (parsed.Project() != pm.Want.Project ||
		parsed.Service() != pm.Want.Service ||
		parsed.Remote() != pm.Want.Remote ||
		parsed.IsRemoteFromHost() != pm.Want.IsRemoteFromHost) {
		t.Errorf("Expected values doesn't match on parsed object: %+v", parsed)
	}
}

func TestParseWithDefaultCustomRemote(t *testing.T) {
	var rl = remotes.List{}

	rl.Set("foo", remotes.Entry{
		Service: "example.com",
	})

	rl.Set("alphabet", remotes.Entry{
		Service: "def.ghi.jkl.mnn.opq.rst.uvw.xyz",
	})

	rl.Set("ip", remotes.Entry{
		Service: "11.22.33.44:5555",
	})

	rl.Set("lcp", remotes.Entry{
		Service: "lfr.cloud",
	})

	rl.Set("local", remotes.Entry{
		Service: "wedeploy.me",
	})

	var c = New(&rl)

	for _, k := range parseMocksWithDefaultCustomRemote {
		testParseWithDefaultCustomRemote(c, k, t)
	}
}

func TestParseUnknownRemoteFlagOnly(t *testing.T) {
	var c = New(&remotes.List{})

	parsed, err := c.Parse(ParseFlags{
		Remote: "cloud",
	})

	if parsed != nil {
		t.Errorf("Expected parsed value to be nil, got %v instead", parsed)
	}

	var wantErr = "remote cloud not found"

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestParseUnknownRemoteFlag(t *testing.T) {
	var c = New(&remotes.List{})

	_, err := c.Parse(ParseFlags{
		Host:   "project",
		Remote: "not-found",
	})

	switch err.(type) {
	case ErrorNotFound:
	default:
		t.Errorf(`Expected error "%v" doesn't match expected type`, err)
	}

	var wantErr = "remote not-found not found"

	if err.Error() != wantErr {
		t.Errorf("Expected error to be %v on parsing, got %v instead", wantErr, err)
	}
}

func TestParseNoMixingProjectAndHost(t *testing.T) {
	var c = CommandFlagFromHost{}
	switch _, err := c.Parse(ParseFlags{
		Host:    "foo.wedeploy.me",
		Project: "foo",
	}); err.(type) {
	case ErrorMultiMode:
	default:
		t.Errorf(`Expected error "%v" doesn't match expected type`, err)
	}
}
