package templates

// NOTICE: Based on Docker's docker/utils/templates/templates.go here
// as of da0ccf8e61e4d5d4005e19fcf0115372f09840bf
// For reference, see:
// https://github.com/docker/docker/blob/master/utils/templates/templates.go
// https://github.com/docker/docker/blob/master/LICENSE

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

var cachedTmpl = map[string]*template.Template{}

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
	var err error
	tmpl, ok := cachedTmpl[format]

	if !ok {
		tmpl, err = parse(format)

		if err != nil {
			return "", errwrap.Wrapf("Template parsing error: {{err}}", err)
		}
	}

	return execute(format, data, tmpl)
}

func execute(format string, data interface{}, tmpl *template.Template) (string, error) {
	var buf bytes.Buffer
	var wr io.Writer = &buf

	if err := tmpl.Execute(wr, data); err != nil {
		return "", errwrap.Wrapf("Can't execute template: {{err}}", err)
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
