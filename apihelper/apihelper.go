package apihelper

import (
	"encoding/json"
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
	haltExitCommand = false

	errStream io.Writer = os.Stderr
)

// Auth a Launchpad request with the global authentication data
func Auth(request *launchpad.Launchpad) {
	var csg = config.Stores["global"]
	var username = csg.Get("username")
	var password = csg.Get("password")
	request.Auth(username, password)
}

// DecodeJSON decodes a JSON response or exits the process on error
func DecodeJSON(request *launchpad.Launchpad, data interface{}) {
	body, err := ioutil.ReadAll(request.Response.Body)

	if err == nil {
		err = json.Unmarshal(body, &data)
	}

	if err != nil {
		fmt.Fprintln(errStream, err)
	}

	printHTTPVerbose(request, body)

	if err != nil {
		exitCommand()
	}
}

// URL creates a Launchpad URL instance
func URL(paths ...string) *launchpad.Launchpad {
	var csg = config.Stores["global"]
	return launchpad.URL(csg.Get("endpoint"), paths...)
}

// ValidateOrExit validates a request or exits the process on error
func ValidateOrExit(request *launchpad.Launchpad, err error) {
	switch err {
	case nil:
		return
	case launchpad.ErrUnexpectedResponse:
		printHTTPError(request)
	default:
		fmt.Fprintln(errStream, err)
	}

	exitCommand()
}

func exitCommand() {
	if !haltExitCommand {
		os.Exit(1)
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
	verbose.Debug(request.Request.Method + " " + request.URL)
	verbose.Debug("Response Body:\n" + string(body))
}
