package apihelper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/fatih/color"
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

func (a APIFault) Error() string {
	return fmt.Sprintf("Launchpad API error: %v", a.Message)
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
	var token = csg.Get("token")

	if token == "" {
		var username = csg.Get("username")
		var password = csg.Get("password")
		request.Auth(username, password)
	} else {
		request.Auth(csg.Get("token"))
	}
}

// DecodeJSON decodes a JSON response
func DecodeJSON(request *launchpad.Launchpad, data interface{}) error {
	body, err := ioutil.ReadAll(request.Response.Body)

	if err == nil {
		err = json.Unmarshal(body, &data)
	}

	return err
}

// DecodeJSONOrExit decodes a JSON response or exits the process on error
func DecodeJSONOrExit(request *launchpad.Launchpad, data interface{}) {
	if err := DecodeJSON(request, data); err != nil {
		fmt.Fprintln(errStream, err)
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
		case string, int, int64, float64, bool:
			request.Param(k, fmt.Sprintf("%v", value))
		default:
			panic(ErrExtractingParams)
		}
	}
}

// RequestVerboseFeedback prints to the verbose err stream info about request
func RequestVerboseFeedback(request *launchpad.Launchpad) {
	if !verbose.Enabled {
		return
	}

	if request.Request == nil {
		verbose.Debug(">", color.RedString("(wait)"), request.URL)
		return
	}

	verbose.Debug(">",
		color.BlueString(request.Request.Method),
		color.YellowString(request.URL),
		color.BlueString(request.Request.Proto))

	for h, r := range request.Headers {
		if len(r) == 1 {
			verbose.Debug(color.BlueString(h)+color.RedString(":"), color.YellowString(r[0]))
		} else {
			verbose.Debug(color.BlueString(h)+color.RedString(":"), r)
		}
	}

	feedbackRequestBody(request)

	verbose.Debug("\n")

	if request.Response == nil {
		verbose.Debug(color.RedString("(null response)"))
		return
	}

	verbose.Debug("<",
		color.BlueString(request.Response.Proto),
		color.RedString(request.Response.Status))

	for h, r := range request.Response.Header {
		if len(r) == 1 {
			verbose.Debug(color.BlueString(h)+color.RedString(":"), color.YellowString(r[0]))
		} else {
			verbose.Debug(color.BlueString(h)+color.RedString(":"), r)
		}
	}

	verbose.Debug("")

	feedbackResponseBody(request.Response)
}

// URL creates a Launchpad URL instance
func URL(paths ...string) *launchpad.Launchpad {
	var csg = config.Stores["global"]
	return launchpad.URL(csg.Get("endpoint"), paths...)
}

// Validate validates a request and sends an error on error
func Validate(request *launchpad.Launchpad, err error) error {
	RequestVerboseFeedback(request)

	switch err {
	case nil:
		return nil
	case launchpad.ErrUnexpectedResponse:
		printHTTPError(request)
	default:
		fmt.Fprintln(errStream, err)
	}

	return err
}

// ValidateOrExit validates a request or exits the process on error
func ValidateOrExit(request *launchpad.Launchpad, err error) {
	err = Validate(request, err)

	if err != nil {
		exitCommand(1)
	}
}

func exitCommand(code int) {
	if !haltExitCommand {
		os.Exit(code)
	}
}

func feedbackRequestBody(request *launchpad.Launchpad) {
	var body = request.RequestBody

	if body != nil {
		verbose.Debug("")
	}

	switch body.(type) {
	case nil:
	case *os.File:
		var fr = body.(*os.File)
		verbose.Debug(
			color.MagentaString("Sending file as request body:\n%v", fr.Name()))
	case *bytes.Buffer:
		verbose.Debug(fmt.Sprintf("\n%s", body.(*bytes.Buffer)))
	case *bytes.Reader:
		var br = body.(*bytes.Reader)
		var b bytes.Buffer
		br.Seek(0, 0)
		br.WriteTo(&b)
		verbose.Debug("\n" + b.String())
	case *strings.Reader:
		var sr = body.(*strings.Reader)
		var b bytes.Buffer
		sr.Seek(0, 0)
		sr.WriteTo(&b)
		verbose.Debug("\n" + (b.String()))
	default:
		verbose.Debug("\n", color.RedString(
			"(request body: "+reflect.TypeOf(body).String()+")"),
		)
	}
}

func feedbackResponseBody(response *http.Response) {
	var body, err = ioutil.ReadAll(response.Body)
	var out bytes.Buffer

	if err != nil {
		verbose.Debug("Error reading response body")
		verbose.Debug(err)
		return
	}

	response.Body = ioutil.NopCloser(bytes.NewReader(body))

	if strings.Contains(
		response.Header.Get("Content-Type"),
		"application/json") {
		err = json.Indent(&out, body, "", "    ")
	}

	if out.Len() == 0 {
		out.Write(body)
	}

	verbose.Debug(color.MagentaString(out.String()) + "\n")
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
}
