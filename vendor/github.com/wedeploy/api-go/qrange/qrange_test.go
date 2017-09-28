package qrange

import (
	"testing"

	"github.com/wedeploy/api-go/jsonlib"
)

func TestFromTo(t *testing.T) {
	var want = `{
		"from": 10,
		"to": 20
}`

	var got = Between(10, 20)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestFromOnly(t *testing.T) {
	var want = `{"from":10}`
	var got = From(10)
	jsonlib.AssertJSONMarshal(t, want, got)
}

func TestToOnly(t *testing.T) {
	var want = `{"to":20}`
	var got = To(20)
	jsonlib.AssertJSONMarshal(t, want, got)
}
