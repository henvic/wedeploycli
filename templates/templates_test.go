// Licensed under Apache license 2.0
// SPDX-License-Identifier: Apache-2.0
// Copyright 2013-2016 Docker, Inc.

// NOTICE: export from moby/utils/templates/templates_test.go (modified)
// https://github.com/moby/moby/blob/da0ccf8e61e4d5d4005e19fcf0115372f09840bf/utils/templates/templates_test.go
// https://github.com/moby/moby/blob/da0ccf8e61e4d5d4005e19fcf0115372f09840bf/LICENSE

package templates

import (
	"testing"
	"text/template"

	"github.com/henvic/wedeploy-sdk-go/jsonlib"
)

type mock struct {
	ID         int
	Name       string
	unexported string
}

func mockf() string {
	return "example"
}

func TestFuncs(t *testing.T) {
	Funcs(template.FuncMap{
		"mockf": mockf,
	})

	var got, err = Execute("{{mockf}}", nil)

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	var want = "example"

	if got != want {
		t.Errorf("Expected value to be \"%v\", got \"%v\" instead", want, got)
	}
}

func TestExecuteOrListExecute(t *testing.T) {
	var want = `"Hello"`
	var got, err = ExecuteOrList(`{{json .Name}}`, mock{104, "Hello", "bar"})

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if want != got {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestExecuteOrListList(t *testing.T) {
	var want = mock{104, "Hello", "not exported; doesn't matter"}

	var got, err = ExecuteOrList("", mock{104, "Hello", "bar"})

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	jsonlib.AssertJSONMarshal(t, got, want)
}

func TestExecuteOrListFailure(t *testing.T) {
	var _, err = ExecuteOrList("", map[interface{}]string{})

	var wantErr = `json: unsupported type: map[interface {}]string`

	if err == nil || err.Error() != wantErr {
		t.Errorf("Expected error to be %v, got %v instead", wantErr, err)
	}
}

func TestExecuteStringFunction(t *testing.T) {
	var want = "text/with/colon"
	var got, err = Execute(`{{join (split . ":") "/"}}`, "text:with:colon")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if want != got {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestExecutePadWithSpaceFunction(t *testing.T) {
	var want = "   Test       "
	var got, err = Execute(`{{- pad . 3 7}}`, "Test")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if want != got {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestExecutePadWithSpaceFunctionEmpty(t *testing.T) {
	var want = ""
	var got, err = Execute(`{{- pad . 3 7}}`, "")

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if want != got {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestExecuteStringFunctionJSON(t *testing.T) {
	var want = `"Hello"`
	var got, err = Execute(`{{json .Name}}`, mock{104, "Hello", "bar"})

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if want != got {
		t.Errorf("Wanted %v, got %v instead", want, got)
	}
}

func TestExecuteCompileError(t *testing.T) {
	var wantErr = `template parsing error: template: :1: unexpected "\\" in command`
	var _, err = Execute("this is a {{ \\ }}", "string")

	if err == nil || err.Error() != wantErr {
		t.Errorf("Wanted err to be %v, got %v instead", wantErr, err)
	}
}

func TestExecuteRunError(t *testing.T) {
	var wantErr = `can't execute template: template: :1:13: executing "" at <.>: can't give argument to non-function .`
	var _, err = Execute("this is a {{ . . }}", 1)

	if err == nil || err.Error() != wantErr {
		t.Errorf("Wanted err to be %v, got %v instead", wantErr, err)
	}
}
