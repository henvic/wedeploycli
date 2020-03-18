package verbosereq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/envs"
	"github.com/henvic/wedeploycli/prettyjson"
	"github.com/henvic/wedeploycli/verbose"
	wedeploy "github.com/wedeploy/wedeploy-sdk-go"
)

// Disabled flag
var Disabled = false

// BlacklistHeadersValues of sensitive information headers
var BlacklistHeadersValues = map[string]bool{
	"Authorization":       true,
	"Set-Cookie":          true,
	"Cookie":              true,
	"Proxy-Authorization": true,
}

var log func(...interface{})

// SetLogFunc for printing the verbosereq debug message
func SetLogFunc(i func(...interface{})) {
	log = i
}

var unsafeVerbose = false

func init() {
	SetLogFunc(verbose.Debug)
	unsafeVerbose = verbose.IsUnsafeMode()
}

// debugRequestBody prints verbose messages when debugging is enabled
func debugRequestBody(body io.Reader) {
	if body != nil {
		log("")
	}

	switch body.(type) {
	case nil:
	case *os.File:
		debugFileReaderBody(body)
	case *bytes.Buffer:
		debugBufferReaderBody(body)
	case *bytes.Reader:
		debugBytesReaderBody(body)
	case *strings.Reader:
		debugStringsReaderBody(body)
	default:
		debugUnknownTypeBody(body)
	}
}

func debugFileReaderBody(body io.Reader) {
	var fr = body.(*os.File)
	log(
		color.Format(color.FgMagenta, "Sending file as request body:\n%v", fr.Name()))
}

func debugBufferReaderBody(body io.Reader) {
	log(fmt.Sprintf("\n%s", body.(*bytes.Buffer)))
}

func debugBytesReaderBody(body io.Reader) {
	var br = body.(*bytes.Reader)
	var b bytes.Buffer

	if _, err := br.Seek(0, 0); err != nil {
		panic(err)
	}

	if _, err := br.WriteTo(&b); err != nil {
		panic(err)
	}

	log("\n" + b.String())
}

func debugStringsReaderBody(body io.Reader) {
	var sr = body.(*strings.Reader)
	var b bytes.Buffer

	if _, err := sr.Seek(0, 0); err != nil {
		panic(err)
	}

	if _, err := sr.WriteTo(&b); err != nil {
		panic(err)
	}

	log("\n" + (b.String()))
}

func debugUnknownTypeBody(body io.Reader) {
	log("\n" + color.Format(
		color.FgRed,
		"(request body: "+reflect.TypeOf(body).String()+")"),
	)
}

// Feedback prints to the verbose err stream info about request
func Feedback(request *wedeploy.WeDeploy) {
	if Disabled || !verbose.Enabled {
		return
	}

	if ctx := request.Context(); ctx != nil && ctx.Value(contextNoVerbose{}) != nil {
		if unsafe, _ := os.LookupEnv(envs.UnsafeVerbose); unsafe != "true" {
			return
		}
	}

	if request.Request == nil {
		log(">", color.Format(color.FgRed, "(wait)"), request.URL)
		return
	}

	requestVerboseFeedback(request)
}

// ContextNoVerbose overrides the verbose.
// The WEDEPLOY_UNSAFE_VERBOSE environment variable overrides this as well.
func ContextNoVerbose(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextNoVerbose{}, true)
}

type contextNoVerbose struct{}

func requestVerboseFeedback(request *wedeploy.WeDeploy) {
	log(">",
		color.Format(color.FgBlue, color.Bold, request.Request.Method),
		color.Format(color.FgYellow, request.URL),
		color.Format(color.FgBlue, request.Request.Proto))

	verbosePrintHeaders(request.Headers)
	debugRequestBody(request.RequestBody)

	log("\n")
	feedbackResponse(request.Response)
}

func feedbackResponse(response *http.Response) {
	if response == nil {
		log(color.Format(color.FgRed, "(null response)"))
		return
	}

	log("<",
		color.Format(color.FgBlue, color.Bold, response.Proto),
		color.Format(color.FgRed, response.Status))

	verbosePrintHeaders(response.Header)
	log("")

	feedbackResponseBody(response)
}

func feedbackResponseBody(response *http.Response) {
	var body, err = ioutil.ReadAll(response.Body)

	if err != nil {
		log("Error reading response body")
		log(err)
		return
	}

	feedbackResponseBodyAll(response, body)
}

func feedbackResponseBodyAll(response *http.Response, body []byte) {
	response.Body = ioutil.NopCloser(bytes.NewReader(body))

	jsonContentType := strings.Contains(response.Header.Get("Content-Type"), "application/json")

	if jsonContentType {
		err := json.Unmarshal(body, &json.RawMessage{})

		if err == nil {
			log(string(prettyjson.Pretty(body)))
			return
		}

		log("Invalid JSON response body:")
		log(err)
	}

	log(color.Format(color.FgMagenta, string(body)+"\n"))
}

func getHeaderValue(key string, values []string) string {
	var v string

	if BlacklistHeadersValues[key] && !unsafeVerbose {
		return verbose.SafeEscapeSlice(values)
	}

	switch len(values) {
	case 0:
		v = color.Format(color.FgRed, "(no values)")
	case 1:
		v = color.Format(color.FgYellow, values[0])
	default:
		v = "[" + color.Format(color.FgYellow, strings.Join(values, " ")) + "]"
	}

	return v
}

func verbosePrintHeaders(headers http.Header) {
	for h, r := range headers {
		log(color.Format(color.FgBlue, color.Bold, h)+
			color.Format(color.FgRed, ":"),
			getHeaderValue(h, r))
	}
}
