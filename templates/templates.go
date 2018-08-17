// Licensed under Apache license 2.0
// SPDX-License-Identifier: Apache-2.0
// Copyright 2013-2016 Docker, Inc.

// NOTICE: export from moby/utils/templates/templates.go (modified)
// https://github.com/moby/moby/blob/da0ccf8e61e4d5d4005e19fcf0115372f09840bf/utils/templates/templates.go
// https://github.com/moby/moby/blob/da0ccf8e61e4d5d4005e19fcf0115372f09840bf/LICENSE

package templates

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"text/template"

	"github.com/hashicorp/errwrap"
)

var basicFunctions = template.FuncMap{
	"json": func(v interface{}) string {
		a, _ := json.Marshal(v)
		return string(a)
	},
	"split": strings.Split,
	"join":  strings.Join,
	"title": strings.Title,
	"lower": strings.ToLower,
	"upper": strings.ToUpper,
	"pad":   padWithSpace,
}

type compiledTemplate struct {
	template *template.Template
	err      error
}

var cachedTmpl = map[string]compiledTemplate{}

// ExecuteOrList executes the Execute function if format is not empty,
// otherwise, it returns all as a JSON
func ExecuteOrList(format string, data interface{}) (string, error) {
	if format != "" {
		return Execute(format, data)
	}

	bin, err := json.MarshalIndent(&data, "", "    ")

	if err != nil {
		return "", err
	}

	return string(bin), nil
}

// Execute template with format and data values
func Execute(format string, data interface{}) (string, error) {
	if _, ok := cachedTmpl[format]; !ok {
		tmpl, tmplErr := parse(format)
		cachedTmpl[format] = compiledTemplate{tmpl, tmplErr}
	}

	return execute(format, data)
}

func execute(format string, data interface{}) (string, error) {
	var (
		buf bytes.Buffer
		wr  io.Writer = &buf

		cached = cachedTmpl[format]
		tmpl   = cached.template
		err    = cached.err
	)

	if err != nil {
		return "", errwrap.Wrapf("template parsing error: {{err}}", err)
	}

	if err := tmpl.Execute(wr, data); err != nil {
		return "", errwrap.Wrapf("can't execute template: {{err}}", err)
	}

	return buf.String(), nil
}

// Parse creates a new annonymous template with the basic functions
// and parses the given format.
func parse(format string) (*template.Template, error) {
	return newParse("", format)
}

// NewParse creates a new tagged template with the basic functions
// and parses the given format.
func newParse(tag, format string) (*template.Template, error) {
	return template.New(tag).Funcs(basicFunctions).Parse(format)
}

// padWithSpace adds whitespace to the input if the input is non-empty
func padWithSpace(source string, prefix, suffix int) string {
	if source == "" {
		return source
	}
	return strings.Repeat(" ", prefix) + source + strings.Repeat(" ", suffix)
}
