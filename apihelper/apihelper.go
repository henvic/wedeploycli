package apihelper

// apihelper SHOULD NOT call any verbose.METHOD directly
// instead, it SHOULD use the verbosereq package
// there is a hidden global debugging flag --no-verbose-requests
// to hide verbose messages related to requests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
)

// APIFault is sent by the server when errors happen
// Method and URL MUST NOT have JSON tags
type APIFault struct {
	Method  string
	URL     string
	Status  int            `json:"status"`
	Message string         `json:"message"`
	Errors  APIFaultErrors `json:"errors"`
}

var errMissingResponse = errors.New("Request is not fulfilled by a response")

func (a APIFault) Error() string {
	return fmt.Sprintf("WeDeploy API error:%s%s%s",
		a.getErrorMessage(),
		a.getErrorURL(),
		a.getErrorMessages())
}

func (a APIFault) getErrorMessage() string {
	var s string

	if a.Status != 0 {
		s += fmt.Sprintf(" %v", a.Status)
	}

	if a.Message != "" {
		s += " " + a.Message
	}

	return s
}

func (a APIFault) getErrorURL() string {
	var s string

	if a.Method != "" || a.URL != "" {
		s += " (" + a.Method + " " + a.URL + ")"
	}

	return s
}

// Get error message for a given reason if it exists
func (a APIFault) Get(reason string) (bool, string) {
	if a.Errors == nil {
		return false, ""
	}

	for _, ed := range a.Errors {
		if ed.Reason == reason {
			return true, ed.Context.Message()
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
		s = append(s, fmt.Sprintf("\n\t%v: %v", value.Context.Message(), value.Reason))
	}

	return strings.Join(s, "")
}

// APIFaultErrors is the array of APIFaultError
type APIFaultErrors []APIFaultError

// APIFaultError is the error structure for the errors described by a fault
type APIFaultError struct {
	Reason  string               `json:"reason"`
	Context APIFaultErrorContext `json:"context"`
}

// APIFaultErrorContext map
type APIFaultErrorContext map[string]string

// Message for a given APIFaultError
func (c APIFaultErrorContext) Message() string {
	if c == nil {
		return ""
	}

	return c["message"]
}

var (
	// ErrInvalidContentType is used when the content-type is not application/json
	ErrInvalidContentType = errors.New("Can only decode data for application/json")

	// ErrExtractingParams is used when query string params fail due to unrecognized type
	ErrExtractingParams = errors.New("Can only extract query string params from flat objects")

	errStream io.Writer = os.Stderr
)

// Auth a WeDeploy request with the global authentication data
func Auth(request *wedeploy.WeDeploy) {
	switch {
	case config.Context.Token == "":
		request.Auth(config.Context.Username, config.Context.Password)
	default:
		request.Auth(config.Context.Token)
	}
}

// AuthGet creates an authenticated GET request for a JSON response end-point
func AuthGet(ctx context.Context, path string, data interface{}) error {
	var request = URL(ctx, path)
	Auth(request)

	if err := Validate(request, request.Get()); err != nil {
		return err
	}

	return DecodeJSON(request, &data)
}

// DecodeJSON decodes a JSON response
func DecodeJSON(request *wedeploy.WeDeploy, data interface{}) error {
	var response = request.Response

	if response == nil {
		return errMissingResponse
	}

	var contentType = response.Header.Get("Content-Type")

	if !strings.Contains(contentType, "application/json") {
		return ErrInvalidContentType
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return errwrap.Wrapf("DecodeJSON error: {{err}}", err)
	}

	err = json.Unmarshal(body, &data)

	if err != nil {
		err = errwrap.Wrapf("Error while decoding JSON: {{err}}", err)
	}

	return err
}

// EncodeJSON encodes a object using its JSON annotations map
// and creates a reader that can be used as body for requests, for example
func EncodeJSON(data interface{}) (*bytes.Reader, error) {
	var b, err = json.MarshalIndent(data, "", "    ")
	return bytes.NewReader(b), err
}

// ParamsFromJSON creates query string params from a flat object with JSON tags
func ParamsFromJSON(request *wedeploy.WeDeploy, data interface{}) {
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

// SetBody sets the body of a request with the JSON encoded from an object
func SetBody(request *wedeploy.WeDeploy, data interface{}) error {
	var r, err = EncodeJSON(&data)

	if err != nil {
		return err
	}

	request.Body(r)
	return err
}

// URL creates a WeDeploy URL instance
func URL(ctx context.Context, paths ...string) *wedeploy.WeDeploy {
	u := wedeploy.URL(config.Context.Infrastructure, paths...)
	u.SetContext(ctx)
	return u
}

// Validate validates a request and sends an error on error
func Validate(request *wedeploy.WeDeploy, err error) error {
	verbosereq.Feedback(request)

	if err == nil {
		return nil
	}

	if ue, ok := err.(*url.Error); ok {
		return handleURLError(ue)
	}

	if err == wedeploy.ErrUnexpectedResponse {
		if af := reportHTTPError(request); af != nil {
			return af
		}
	}

	return err
}

func handleURLError(ue *url.Error) error {
	if config.Context == nil || config.Global == nil {
		return ue
	}

	// only show this error if
	// the active infrastructure domain is related to...
	var s = "WeDeploy platform error:"

	switch {
	case verbose.Enabled:
		s += "\n{{err}}"
	case config.Context.Remote == defaults.LocalRemote:
		s += " local infrastructure is not running\n"
		s += `tip: try running "we deploy --start-local-infra" and then run this command again`
	default:
		s += " could not connect to remote infrastructure"
	}

	return errwrap.Wrapf(s, ue)
}

func reportHTTPError(request *wedeploy.WeDeploy) error {
	var body, err = ioutil.ReadAll(request.Response.Body)

	if err != nil {
		return err
	}

	var jsonErr bool
	jsonErr, err = reportHTTPErrorTryJSON(request, body)

	if err == nil || jsonErr {
		return reportHTTPErrorNotJSON(request, body)
	}

	return err
}

func reportHTTPErrorTryJSON(request *wedeploy.WeDeploy, body []byte) (bool, error) {
	var response = request.Response
	var contentType = response.Header.Get("Content-Type")
	var af APIFault

	if !strings.Contains(contentType, "application/json") {
		return true, nil
	}

	if ed := json.Unmarshal(body, &af); ed != nil {
		return true, ed
	}

	return false, reportHTTPErrorJSON(request, af)
}

func reportHTTPErrorJSON(request *wedeploy.WeDeploy, af APIFault) error {
	af.Method = request.Request.Method
	af.URL = request.URL
	return &af
}

func reportHTTPErrorNotJSON(
	request *wedeploy.WeDeploy, body []byte) *APIFault {
	var response = request.Response
	var fault = &APIFault{
		Method:  request.Request.Method,
		URL:     request.URL,
		Status:  response.StatusCode,
		Message: http.StatusText(response.StatusCode),
		Errors:  APIFaultErrors{},
	}

	if len(body) != 0 {
		fault.Errors = append(fault.Errors, APIFaultError{
			Reason: string(body),
			Context: APIFaultErrorContext{
				"message": "Response Body is not JSON",
			},
		})
	}

	return fault
}
