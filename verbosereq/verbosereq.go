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

	"github.com/fatih/color"
	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/verbose"
)

// DebugRequestBody prints verbose messages when debugging is enabled
func DebugRequestBody(body io.Reader) {
	if body != nil {
		verbose.Debug("")
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
	verbose.Debug(
		color.MagentaString("Sending file as request body:\n%v", fr.Name()))
}

func debugBufferReaderBody(body io.Reader) {
	verbose.Debug(fmt.Sprintf("\n%s", body.(*bytes.Buffer)))
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

	verbose.Debug("\n" + b.String())
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

	verbose.Debug("\n" + (b.String()))
}

func debugUnknownTypeBody(body io.Reader) {
	verbose.Debug("\n" + color.RedString(
		"(request body: "+reflect.TypeOf(body).String()+")"),
	)
}

// Feedback prints to the verbose err stream info about request
func Feedback(request *wedeploy.WeDeploy) {
	if !verbose.Enabled {
		return
	}

	if request.Request == nil {
		verbose.Debug(">", color.RedString("(wait)"), request.URL)
		return
	}

	requestVerboseFeedback(request)
}

func requestVerboseFeedback(request *wedeploy.WeDeploy) {
	verbose.Debug(">",
		color.BlueString(request.Request.Method),
		color.YellowString(request.URL),
		color.BlueString(request.Request.Proto))

	verbosePrintHeaders(request.Headers)
	DebugRequestBody(request.RequestBody)

	verbose.Debug("\n")
	feedbackResponse(request.Response)
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

	if err != nil {
		verbose.Debug("Error reading response body")
		verbose.Debug(err)
		return
	}

	feedbackResponseBodyAll(response, body)
}

func feedbackResponseBodyReadJSON(response *http.Response, body []byte) (
	out bytes.Buffer) {
	if strings.Contains(
		response.Header.Get("Content-Type"), "application/json") {
		if err := json.Indent(&out, body, "", "    "); err != nil {
			verbose.Debug("Response not JSON (as expected by Content-Type)")
			verbose.Debug(err)
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

	verbose.Debug(color.MagentaString(out.String()) + "\n")
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
