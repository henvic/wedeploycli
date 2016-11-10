package templates

// NOTICE: Based on Docker's docker/utils/templates/templates.go here
// as of da0ccf8e61e4d5d4005e19fcf0115372f09840bf
// For reference, see:
// https://github.com/docker/docker/blob/master/utils/templates/templates_test.go
// https://github.com/docker/docker/blob/master/LICENSE

import (
	"bytes"
	"testing"
)

func TestParseStringFunctions(t *testing.T) {
	tm, err := parse(`{{join (split . ":") "/"}}`)
	if err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	if err := tm.Execute(&b, "text:with:colon"); err != nil {
		t.Fatal(err)
	}
	want := "text/with/colon"
	if b.String() != want {
		t.Fatalf("expected %s, got %s", want, b.String())
	}
}

func TestNewParse(t *testing.T) {
	tm, err := newParse("foo", "this is a {{ . }}")
	if err != nil {
		t.Fatal(err)
	}

	var b bytes.Buffer
	if err := tm.Execute(&b, "string"); err != nil {
		t.Fatal(err)
	}
	want := "this is a string"
	if b.String() != want {
		t.Fatalf("expected %s, got %s", want, b.String())
	}
}
