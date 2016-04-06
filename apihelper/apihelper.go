package apihelper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/verbose"
)

// APIFault is sent by the server when errors happen
type APIFault struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Errors  []APIFaultError `json:"errors"`
}

// APIFaultError is the error structure for the errors described by a fault
type APIFaultError struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

var (
	// ErrExtractingParams is used when query string params fail due to unrecognized type
	ErrExtractingParams = errors.New("Can only extract query string params from flat objects")

	errStream       io.Writer = os.Stderr
	haltExitCommand           = false
)

// Auth a Launchpad request with the global authentication data
func Auth(request *launchpad.Launchpad) {
	var csg = config.Stores["global"]
	var username = csg.Get("username")
	var password = csg.Get("password")
	request.Auth(username, password)
}

// DecodeJSON decodes a JSON response
func DecodeJSON(request *launchpad.Launchpad, data interface{}) error {
	body, err := ioutil.ReadAll(request.Response.Body)

	if err == nil {
		err = json.Unmarshal(body, &data)
	}

	if err != nil {
		fmt.Fprintln(errStream, err)
	}

	printHTTPVerbose(request, body)

	return err
}

// DecodeJSONOrExit decodes a JSON response or exits the process on error
func DecodeJSONOrExit(request *launchpad.Launchpad, data interface{}) {
	if err := DecodeJSON(request, data); err != nil {
		exitCommand(1)
	}
}

// ParamsFromJSON creates query string params from a flat object with JSON tags
func ParamsFromJSON(request *launchpad.Launchpad, data interface{}) {
	var v map[string]interface{}

	b, err := json.Marshal(data)

	if err == nil {
		err = json.Unmarshal(b, &v)
	}

	if err != nil {
		panic(err)
	}

	for k, value := range v {
		switch value.(type) {
		case nil:
			request.Param(k, "null")
		case string, int, int64, float64:
			request.Param(k, fmt.Sprintf("%v", value))
		default:
			panic(ErrExtractingParams)
		}
	}
}

// URL creates a Launchpad URL instance
func URL(paths ...string) *launchpad.Launchpad {
	var csg = config.Stores["global"]
	return launchpad.URL(csg.Get("endpoint"), paths...)
}

// ValidateOrExit validates a request or exits the process on error
func ValidateOrExit(request *launchpad.Launchpad, err error) {

	if request.Request == nil {
		verbose.Debug("(wait) " + request.URL)
	} else {
		verbose.Debug(request.Request.Method + " " + request.URL)
	}

	switch err {
	case nil:
		return
	case launchpad.ErrUnexpectedResponse:
		printHTTPError(request)
	default:
		fmt.Fprintln(errStream, err)
	}

	exitCommand(1)
}

func exitCommand(code int) {
	if !haltExitCommand {
		os.Exit(code)
	}
}

func printErrorList(list []APIFaultError) {
	if list != nil {
		for _, value := range list {
			fmt.Fprintln(errStream, value.Message)
		}
	}
}

func printHTTPError(request *launchpad.Launchpad) {
	var af APIFault

	fmt.Fprintln(errStream, request.Response.Status)

	body, err := ioutil.ReadAll(request.Response.Body)

	if err == nil {
		err = json.Unmarshal(body, &af)
	}

	if err != nil {
		fmt.Fprintln(errStream, string(body))
		return
	}

	printErrorList(af.Errors)
	printHTTPVerbose(request, body)
}

func printHTTPVerbose(request *launchpad.Launchpad, body []byte) {
	verbose.Debug("Response Body:\n" + string(body))
}
