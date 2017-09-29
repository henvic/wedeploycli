package verbosereq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/wedeploy/api-go"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/stringlib"
	"github.com/wedeploy/cli/tdata"
	"github.com/wedeploy/cli/verbose"
)

var (
	bufErrStream bytes.Buffer
)

func TestMain(m *testing.M) {
	var defaultVerboseEnabled = verbose.Enabled
	var defaultVerboseErrStream = verbose.ErrStream
	var defaultNoColor = color.NoColor
	color.NoColor = true
	verbose.Enabled = true
	verbose.ErrStream = &bufErrStream
	ec := m.Run()
	verbose.Enabled = defaultVerboseEnabled
	verbose.ErrStream = defaultVerboseErrStream
	color.NoColor = defaultNoColor
	os.Exit(ec)
}

func TestDebugRequestBody(t *testing.T) {
	bufErrStream.Reset()
	debugRequestBody(nil)

	if bufErrStream.Len() != 0 {
		t.Errorf("Wanted debug to be empty")
	}
}

func TestRequestVerboseFeedback(t *testing.T) {
	bufErrStream.Reset()
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Test-Multiple", "a")
		w.Header().Add("X-Test-Multiple", "b")
		fmt.Fprintf(w, "Hello")
	})

	var request = wedeploy.URL("http://www.example.com/foo")

	request.Headers.Add("Accept", "application/json")
	request.Headers.Add("Accept", "text/plain")

	if err := request.Get(); err != nil {
		panic(err)
	}

	Feedback(request)

	var got = bufErrStream.String()

	var find = []string{
		"> GET http://www.example.com/foo HTTP/1.1",
		"Content-Type: text/plain; charset=utf-8",
		"Accept: [application/json text/plain]",
		"X-Test-Multiple: [a b]",
		"Hello",
	}

	var assertionError = false

	for _, want := range find {
		if !strings.Contains(got, want) {
			assertionError = true
			t.Errorf("Response doesn't contain expected value %v", want)
		}
	}

	if assertionError {
		t.Errorf("Response is:\n%v", got)
	}

	servertest.Teardown()
}

func TestRequestVerboseFeedbackOff(t *testing.T) {
	bufErrStream.Reset()
	servertest.Setup()
	verbose.Enabled = false

	servertest.Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
	})

	var request = wedeploy.URL("http://www.example.com/foo")

	if err := request.Get(); err != nil {
		panic(err)
	}

	Feedback(request)

	var got = bufErrStream.String()

	if len(got) != 0 {
		t.Errorf("Expected to generate no output.")
	}

	servertest.Teardown()
	verbose.Enabled = true
}

func TestRequestVerboseFeedbackUpload(t *testing.T) {
	bufErrStream.Reset()
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo", tdata.ServerHandler(""))

	var request = wedeploy.URL("http://www.example.com/foo")

	var file, err = os.Open("mocks/config.json")

	if err != nil {
		panic(err)
	}

	request.Body(file)

	if err := request.Get(); err != nil {
		panic(err)
	}

	Feedback(request)

	var got = bufErrStream.String()

	var find = []string{
		"> GET http://www.example.com/foo HTTP/1.1",
		"Sending file as request body:\nmocks/config.json",
	}

	var assertionError = false

	for _, want := range find {
		if !strings.Contains(got, want) {
			assertionError = true
			t.Errorf("Response doesn't contain expected value %v", want)
		}
	}

	if assertionError {
		t.Errorf("Response is:\n%v", got)
	}

	servertest.Teardown()
}

func TestRequestVerboseFeedbackStringReader(t *testing.T) {
	bufErrStream.Reset()
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo", tdata.ServerHandler(""))

	var request = wedeploy.URL("http://www.example.com/foo")

	request.Body(strings.NewReader("custom body"))

	if err := request.Get(); err != nil {
		panic(err)
	}

	Feedback(request)

	var got = bufErrStream.String()

	var find = []string{
		"> GET http://www.example.com/foo HTTP/1.1",
		"\ncustom body\n",
	}

	var assertionError = false

	for _, want := range find {
		if !strings.Contains(got, want) {
			assertionError = true
			t.Errorf("Response doesn't contain expected value %v", want)
		}
	}

	if assertionError {
		t.Errorf("Response is:\n%v", got)
	}

	servertest.Teardown()
}

func TestRequestVerboseFeedbackBytesReader(t *testing.T) {
	bufErrStream.Reset()
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo", tdata.ServerHandler(""))

	var request = wedeploy.URL("http://www.example.com/foo")

	var sr = strings.NewReader("custom body")

	var b bytes.Buffer

	if _, err := sr.WriteTo(&b); err != nil {
		panic(err)
	}

	request.Body(bytes.NewReader(b.Bytes()))

	if err := request.Get(); err != nil {
		panic(err)
	}

	Feedback(request)

	var got = bufErrStream.String()

	var find = []string{
		"> GET http://www.example.com/foo HTTP/1.1",
		"\ncustom body\n",
	}

	var assertionError = false

	for _, want := range find {
		if !strings.Contains(got, want) {
			assertionError = true
			t.Errorf("Response doesn't contain expected value %v", want)
		}
	}

	if assertionError {
		t.Errorf("Response is:\n%v", got)
	}

	servertest.Teardown()
}

func TestRequestVerboseFeedbackOtherReader(t *testing.T) {
	bufErrStream.Reset()
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo", tdata.ServerHandler(""))

	var request = wedeploy.URL("http://www.example.com/foo")

	var body = strings.NewReader("custom body")

	var b = &bytes.Buffer{}

	request.Body(io.TeeReader(body, b))

	if err := request.Get(); err != nil {
		panic(err)
	}

	Feedback(request)

	var got = bufErrStream.String()

	var find = []string{
		"> GET http://www.example.com/foo HTTP/1.1",
		"\n(request body: *io.teeReader)\n",
	}

	var assertionError = false

	for _, want := range find {
		if !strings.Contains(got, want) {
			assertionError = true
			t.Errorf("Response doesn't contain expected value %v", want)
		}
	}

	if assertionError {
		t.Errorf("Response is:\n%v", got)
	}

	servertest.Teardown()
}

func TestRequestVerboseFeedbackJSONResponse(t *testing.T) {
	bufErrStream.Reset()
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"Hello": "World"}`)
	})

	var request = wedeploy.URL("http://www.example.com/foo")

	type Foo struct {
		Bar string `json:"bar"`
	}

	var foo = &Foo{Bar: "one"}

	var b, err = json.Marshal(foo)

	if err != nil {
		panic(err)
	}

	request.Body(bytes.NewBuffer(b))

	if err := request.Post(); err != nil {
		panic(err)
	}

	Feedback(request)

	var got = bufErrStream.String()

	var find = []string{
		"> POST http://www.example.com/foo HTTP/1.1",
		`{"bar":"one"}`,
		"{\n    \"Hello\": \"World\"\n}",
	}

	var assertionError = false

	for _, want := range find {
		if !strings.Contains(got, want) {
			assertionError = true
			t.Errorf("Response doesn't contain expected value %v", want)
		}
	}

	if assertionError {
		t.Errorf("Response is:\n%v", got)
	}

	servertest.Teardown()
}

func TestRequestVerboseFeedbackJSONResponseWithBlacklistedHeaders(t *testing.T) {
	bufErrStream.Reset()
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "Bearer 1")
		w.Header().Add("Set-Cookie", "foo4=bar4")
		w.Header().Add("Cookie", "foo1=bar1")
		w.Header().Add("Cookie", "foo2=bar2")
		w.Header().Set("Proxy-Authorization", "foo")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"Hello": "World"}`)
	})

	var request = wedeploy.URL("http://www.example.com/foo")

	type Foo struct {
		Bar string `json:"bar"`
	}

	var foo = &Foo{Bar: "one"}

	var b, err = json.Marshal(foo)

	if err != nil {
		panic(err)
	}

	request.Body(bytes.NewBuffer(b))

	if err := request.Post(); err != nil {
		panic(err)
	}

	Feedback(request)

	var got = bufErrStream.String()

	var find = []string{
		"> POST http://www.example.com/foo HTTP/1.1",
		`{"bar":"one"}`,
		"{\n    \"Hello\": \"World\"\n}",
		"Authorization:  1 hidden value",
		"Cookie:  1 hidden value",
		"Cookie:  2 hidden values",
		"Proxy-Authorization:  1 hidden value",
	}

	var assertionError = false

	for _, want := range find {
		if !strings.Contains(got, want) {
			assertionError = true
			t.Errorf("Response doesn't contain expected value %v", want)
		}
	}

	if assertionError {
		t.Errorf("Response is:\n%v", got)
	}

	servertest.Teardown()
}

func TestRequestVerboseFeedbackJSONResponseWithBlacklistedHeadersUnsafe(t *testing.T) {
	defer func() {
		unsafeVerbose = false
	}()

	unsafeVerbose = true

	bufErrStream.Reset()
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "Bearer 1")
		w.Header().Add("Set-Cookie", "foo4=bar4")
		w.Header().Add("Cookie", "foo1=bar1")
		w.Header().Add("Cookie", "foo2=bar2")
		w.Header().Set("Proxy-Authorization", "foo")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"Hello": "World"}`)
	})

	var request = wedeploy.URL("http://www.example.com/foo")

	type Foo struct {
		Bar string `json:"bar"`
	}

	var foo = &Foo{Bar: "one"}

	var b, err = json.Marshal(foo)

	if err != nil {
		panic(err)
	}

	request.Body(bytes.NewBuffer(b))

	if err := request.Post(); err != nil {
		panic(err)
	}

	Feedback(request)

	var got = bufErrStream.String()

	var find = []string{
		"> POST http://www.example.com/foo HTTP/1.1",
		`{"bar":"one"}`,
		"{\n    \"Hello\": \"World\"\n}",
		"Authorization: Bearer 1",
		"Cookie: foo4=bar4",
		"Cookie: [foo1=bar1 foo2=bar2]",
		"Proxy-Authorization: foo",
	}

	var assertionError = false

	for _, want := range find {
		if !strings.Contains(got, want) {
			assertionError = true
			t.Errorf("Response doesn't contain expected value %v", want)
		}
	}

	if assertionError {
		t.Errorf("Response is:\n%v", got)
	}

	servertest.Teardown()
}

func TestRequestVerboseFeedbackInvalidJSONResponse(t *testing.T) {
	bufErrStream.Reset()
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"Hello": "World!"`)
	})

	var request = wedeploy.URL("http://www.example.com/foo")

	type Foo struct {
		Bar string `json:"bar"`
	}

	var foo = &Foo{Bar: "one"}

	var b, err = json.Marshal(foo)

	if err != nil {
		panic(err)
	}

	request.Body(bytes.NewBuffer(b))

	if err := request.Post(); err != nil {
		panic(err)
	}

	Feedback(request)

	var got = bufErrStream.String()

	var find = []string{
		"> POST http://www.example.com/foo HTTP/1.1",
		`{"bar":"one"}`,
		"Response not JSON (as expected by Content-Type)",
		"unexpected end of JSON input",
		`{"Hello": "World!"`,
	}

	var assertionError = false

	for _, want := range find {
		if !strings.Contains(got, want) {
			assertionError = true
			t.Errorf("Response doesn't contain expected value %v", want)
		}
	}

	if assertionError {
		t.Errorf("Response is:\n%v", got)
	}

	servertest.Teardown()
}

func TestRequestVerboseFeedbackNullResponse(t *testing.T) {
	bufErrStream.Reset()

	var request = wedeploy.URL("http://www.example.com/foo")

	request.URL = "x://"

	// this returns an error, but we are not going to shortcut to avoid getting
	// the error value here because we want to see what verbose returns
	if err := request.Get(); err == nil {
		t.Errorf("Expected error, got %v instead", err)
	}

	Feedback(request)

	var got = bufErrStream.String()

	var find = []string{
		"> GET x:// HTTP/1.1",
		"(null response)",
	}

	var assertionError = false

	for _, want := range find {
		if !strings.Contains(got, want) {
			assertionError = true
			t.Errorf("Response doesn't contain expected value %v", want)
		}
	}

	if assertionError {
		t.Errorf("Response is:\n%v", got)
	}
}

func TestRequestVerboseFeedbackNotComplete(t *testing.T) {
	bufErrStream.Reset()

	var request = wedeploy.URL("http://www.example.com/foo")

	Feedback(request)

	stringlib.AssertSimilar(t,
		"> (wait) http://www.example.com/foo",
		bufErrStream.String())
}
