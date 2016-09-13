package verbosereq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/verbose"
)

// Disabled flag
var Disabled = false

var log func(...interface{})

// SetLogFunc for printing the verbosereq debug message
func SetLogFunc(i func(...interface{})) {
	log = i
}

func init() {
	SetLogFunc(verbose.Debug)
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

	if request.Request == nil {
		log(">", color.Format(color.FgRed, "(wait)"), request.URL)
		return
	}

	requestVerboseFeedback(request)
}

func requestVerboseFeedback(request *wedeploy.WeDeploy) {
	log(">",
		color.Format(color.FgBlue, request.Request.Method),
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
		color.Format(color.FgBlue, response.Proto),
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

func feedbackResponseBodyReadJSON(response *http.Response, body []byte) (
	out bytes.Buffer) {
	if strings.Contains(
		response.Header.Get("Content-Type"), "application/json") {
		if err := json.Indent(&out, body, "", "    "); err != nil {
			log("Response not JSON (as expected by Content-Type)")
			log(err)
		}
	}

	return out
}

func feedbackResponseBodyAll(response *http.Response, body []byte) {
	response.Body = ioutil.NopCloser(bytes.NewReader(body))
	var out = feedbackResponseBodyReadJSON(response, body)

	if out.Len() == 0 {
		if _, err := out.Write(body); err != nil {
			panic(err)
		}
	}

	log(color.Format(color.FgMagenta, out.String()) + "\n")
}

func verbosePrintHeaders(headers http.Header) {
	for h, r := range headers {
		if len(r) == 1 {
			log(color.Format(color.FgBlue, h)+color.Format(color.FgRed, ":"), color.Format(color.FgYellow, r[0]))
		} else {
			log(color.Format(color.FgBlue, h)+color.Format(color.FgRed, ":"), r)
		}
	}
}
