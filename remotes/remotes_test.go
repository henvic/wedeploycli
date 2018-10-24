package remotes

import (
	"fmt"
	"testing"
)

func TestKeysIsSorted(t *testing.T) {
	var list = List{}

	list.entries = map[string]Entry{
		"alternative": Entry{
			// probably want to remove the "http:// everywhere on this file"
			Infrastructure: "http://example.net/",
			Service:        "foobar.com",
		},
		"staging": Entry{
			Infrastructure: "http://staging.example.net/",
		},
		"beta": Entry{
			Infrastructure:        "http://beta.example.com/",
			InfrastructureComment: "my beta comment",
		},
		"remain": Entry{
			Infrastructure: "http://localhost/",
			Comment:        "commented vars remains even when empty",
		},
		"dontremain": Entry{
			Infrastructure: "http://localhost/",
		},
		"dontremain2": Entry{
			Infrastructure: "http://localhost/",
		},
	}

	var keys = list.Keys()
	var sorted = "[alternative beta dontremain dontremain2 remain staging]"

	if fmt.Sprintf("%v", keys) != sorted {
		t.Errorf("Expected list to be sorted by string, got %v instead", keys)
	}
}

func TestGetAndDelete(t *testing.T) {
	var list = List{}

	list.entries = map[string]Entry{
		"alternative": Entry{
			Infrastructure:        "http://example.net/",
			Comment:               "123",
			InfrastructureComment: "abc",
		},
		"staging": Entry{
			Infrastructure: "http://staging.example.net/",
		},
	}

	alt, ok := list.entries["alternative"]

	if !ok || alt.Infrastructure != "http://example.net/" || alt.Comment != "123" || alt.InfrastructureComment != "abc" {
		t.Errorf("Expected values ((http://example.net/, 123, abc), true), got (%v, %v) instead", alt, ok)
	}

	list.Del("staging")

	if s, ok := list.entries["staging"]; ok {
		t.Errorf(`Expecting "staging" to not exist, got %v instead`, s)
	}
}

func TestGet(t *testing.T) {
	var list = List{}

	list.Set("alternative", Entry{
		Infrastructure: "http://example.net/",
		Comment:        "123",
	})

	if got := list.Get("alternative"); got.Infrastructure != "http://example.net/" || got.Comment != "123" {
		t.Errorf("Expected values (http://example.net/, 123), got %v instead", got)
	}

	if has := list.Has("alternative"); !has {
		t.Error("Expected remote to be found")
	}
}

func TestHasNil(t *testing.T) {
	var list = List{}

	if got := list.Has("foo"); got {
		t.Errorf("Expected object to not exist, got %v instead", got)
	}
}

func TestSet(t *testing.T) {
	var list = List{}

	list.Set("alternative", Entry{
		Infrastructure: "http://example.net/",
		Comment:        "123",
	})

	alt, ok := list.entries["alternative"]

	if !ok || alt.Infrastructure != "http://example.net/" || alt.Comment != "123" {
		t.Errorf("Expected values ((http://example.net/, 123), true), got (%v, %v) instead", alt, ok)
	}
}
