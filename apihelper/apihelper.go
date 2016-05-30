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
// Method and URL MUST NOT have JSON tags
type APIFault struct {
	Method  string
	URL     string
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Errors  APIFaultErrors `json:"errors"`
}

var errInvalidGlobalConfig = errors.New("Can't get endpoint: global config is null")

func (a APIFault) Error() string {
	var s = "WeDeploy API error:"

	if a.Code != 0 {
		s += fmt.Sprintf(" %v", a.Code)
	}

	if a.Message != "" {
		s += " " + a.Message
	}

	if a.Method != "" || a.URL != "" {
		s += " (" + a.Method + " " + a.URL + ")"
	}

	s += a.getErrorMessages()

	return s
}

// Get error message for a given reason if it exists
func (a APIFault) Get(reason string) (bool, string) {
	if a.Errors == nil {
		return false, ""
	}

	for _, ed := range a.Errors {
		if ed.Reason == reason {
			return true, ed.Message
		}
	}

	return false, ""
}

// Has checks if given error reason exists
func (a APIFault) Has(reason string) bool {
	var has, _ = a.Get(reason)
	return has
}

func (a APIFault) getErrorMessages() string {
	var s []string

	if a.Errors == nil {
		return ""
	}

	for _, value := range a.Errors {
		s = append(s, fmt.Sprintf("\n\t%v: %v", value.Message, value.Reason))
	}

	return strings.Join(s, "")
}

// APIFaultErrors is the array of APIFaultError
type APIFaultErrors []APIFaultError

// APIFaultError is the error structure for the errors described by a fault
type APIFaultError struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

var (
	// ErrInvalidContentType is used when the content-type is not application/json
	ErrInvalidContentType = errors.New("Can only decode data for application/json")

	// ErrExtractingParams is used when query string params fail due to unrecognized type
	ErrExtractingParams = errors.New("Can only extract query string params from flat objects")

	errStream       io.Writer = os.Stderr
	haltExitCommand           = false
)

// Auth a WeDeploy request with the global authentication data
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

// AuthGet creates an authenticated GET request for a JSON response end-point
func AuthGet(path string, data interface{}) error {
	var request = URL(path)

	Auth(request)

	if err := Validate(request, request.Get()); err != nil {
		return err
	}

	return DecodeJSON(request, &data)
}

// AuthGetOrExit is like AuthGet, but exits on failure
func AuthGetOrExit(path string, data interface{}) {
	var err = AuthGet(path, data)

	if err != nil {
		exitError(err)
	}
}

// DecodeJSON decodes a JSON response
func DecodeJSON(request *launchpad.Launchpad, data interface{}) error {
	var response = request.Response
	var contentType = response.Header.Get("Content-Type")

	if !strings.Contains(contentType, "application/json") {
		return ErrInvalidContentType
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return err
	}

	return json.Unmarshal(body, &data)
}

// DecodeJSONOrExit decodes a JSON response or exits the process on error
func DecodeJSONOrExit(request *launchpad.Launchpad, data interface{}) {
	if err := DecodeJSON(request, data); err != nil {
		exitError(err)
	}
}

// EncodeJSON encodes a object using its JSON annotations map
// and creates a reader that can be used as body for requests, for example
func EncodeJSON(data interface{}) (*bytes.Reader, error) {
	var b, err = json.Marshal(data)
	return bytes.NewReader(b), err
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

	verbosePrintHeaders(request.Headers)
	feedbackRequestBody(request)

	verbose.Debug("\n")
	feedbackResponse(request.Response)
}

// SetBody sets the body of a request with the JSON encoded from an object
func SetBody(request *launchpad.Launchpad, data interface{}) error {
	var r, err = EncodeJSON(&data)

	if err != nil {
		return err
	}

	request.Body(r)
	return err
}

// URL creates a WeDeploy URL instance
func URL(paths ...string) *launchpad.Launchpad {
	var csg = config.Stores["global"]

	if csg == nil {
		panic(errInvalidGlobalConfig)
	}

	return launchpad.URL(csg.Get("endpoint"), paths...)
}

// Validate validates a request and sends an error on error
func Validate(request *launchpad.Launchpad, err error) error {
	RequestVerboseFeedback(request)

	if err == nil {
		return nil
	}

	if err == launchpad.ErrUnexpectedResponse {
		if af := reportHTTPError(request); af != nil {
			return af
		}
	}

	return err
}

// ValidateOrExit validates a request or exits the process on error
func ValidateOrExit(request *launchpad.Launchpad, err error) {
	err = Validate(request, err)

	if err != nil {
		exitError(err)
	}
}

func exitError(err error) {
	fmt.Fprintln(errStream, err)
	exitCommand(1)
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
		verbose.Debug("\n" + color.RedString(
			"(request body: "+reflect.TypeOf(body).String()+")"),
		)
	}
}

func feedbackResponse(response *http.Response) {
	if response == nil {
		verbose.Debug(color.RedString("(null response)"))
		return
	}

	verbose.Debug("<",
		color.BlueString(response.Proto),
		color.RedString(response.Status))

	verbosePrintHeaders(response.Header)
	verbose.Debug("")

	feedbackResponseBody(response)
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
		if err := json.Indent(&out, body, "", "    "); err != nil {
			verbose.Debug("Response not JSON (as expected by Content-Type)")
			verbose.Debug(err)
		}
	}

	if out.Len() == 0 {
		out.Write(body)
	}

	verbose.Debug(color.MagentaString(out.String()) + "\n")
}

func reportHTTPError(request *launchpad.Launchpad) error {
	var af APIFault

	var response = request.Response
	var contentType = response.Header.Get("Content-Type")
	var body, err = ioutil.ReadAll(response.Body)

	if err != nil {
		return err
	}

	if strings.Contains(contentType, "application/json") {
		err = json.Unmarshal(body, &af)

		if err == nil {
			af.Method = request.Request.Method
			af.URL = request.URL
			return &af
		}

		fmt.Fprintf(errStream, "Failure decoding JSON error: %v", err)
	}

	af = APIFault{
		Method:  request.Request.Method,
		URL:     request.URL,
		Code:    response.StatusCode,
		Message: http.StatusText(response.StatusCode),
		Errors: APIFaultErrors{
			APIFaultError{
				Reason:  string(body),
				Message: "body",
			},
		},
	}

	return &af
}

func verbosePrintHeaders(headers http.Header) {
	for h, r := range headers {
		if len(r) == 1 {
			verbose.Debug(color.BlueString(h)+color.RedString(":"), color.YellowString(r[0]))
		} else {
			verbose.Debug(color.BlueString(h)+color.RedString(":"), r)
		}
	}
}
