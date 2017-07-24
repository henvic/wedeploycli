package remotes

import (
	"fmt"
	"testing"
)

func TestKeysIsSorted(t *testing.T) {
	var list = List{
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
	var list = List{
		"alternative": Entry{
			Infrastructure:        "http://example.net/",
			Comment:               "123",
			InfrastructureComment: "abc",
		},
		"staging": Entry{
			Infrastructure: "http://staging.example.net/",
		},
	}

	alt, ok := list["alternative"]

	if !ok || alt.Infrastructure != "http://example.net/" || alt.Comment != "123" || alt.InfrastructureComment != "abc" {
		t.Errorf("Expected values ((http://example.net/, 123, abc), true), got (%v, %v) instead", alt, ok)
	}

	list.Del("staging")

	if s, ok := list["staging"]; ok {
		t.Errorf(`Expecting "staging" to not exist, got %v instead`, s)
	}
}

func TestSet(t *testing.T) {
	var list = List{}

	list.Set("alternative", Entry{
		Infrastructure: "http://example.net/",
		Comment:        "123",
	})

	alt, ok := list["alternative"]

	if !ok || alt.Infrastructure != "http://example.net/" || alt.Comment != "123" {
		t.Errorf("Expected values ((http://example.net/, 123), true), got (%v, %v) instead", alt, ok)
	}
}
