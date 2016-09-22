package remotes

import (
	"fmt"
	"testing"
)

func TestKeysIsSorted(t *testing.T) {
	var list = List{
		"alternative": Entry{
			URL: "http://example.net/",
		},
		"staging": Entry{
			URL: "http://staging.example.net/",
		},
		"beta": Entry{
			URL:        "http://beta.example.com/",
			URLComment: "my beta comment",
		},
		"remain": Entry{
			URL:     "http://localhost/",
			Comment: "commented vars remains even when empty",
		},
		"dontremain": Entry{
			URL: "http://localhost/",
		},
		"dontremain2": Entry{
			URL: "http://localhost/",
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
			URL:        "http://example.net/",
			Comment:    "123",
			URLComment: "abc",
		},
		"staging": Entry{
			URL: "http://staging.example.net/",
		},
	}

	alt, ok := list.Get("alternative")

	if !ok || alt.URL != "http://example.net/" || alt.Comment != "123" || alt.URLComment != "abc" {
		t.Errorf("Expected values ((http://example.net/, 123, abc), true), got (%v, %v) instead", alt, ok)
	}

	list.Del("staging")

	if s, ok := list.Get("staging"); ok {
		t.Errorf(`Expecting "staging" to not exist, got %v instead`, s)
	}
}

func TestSet(t *testing.T) {
	var list = List{}

	list.Set("alternative", "http://example.net/", "123")

	alt, ok := list.Get("alternative")

	if !ok || alt.URL != "http://example.net/" || alt.Comment != "# 123" {
		t.Errorf("Expected values ((http://example.net/, # 123), true), got (%v, %v) instead", alt, ok)
	}
}
