package apihelper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/kylelemons/godebug/pretty"
	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/globalconfigmock"
	"github.com/launchpad-project/cli/servertest"
	"github.com/launchpad-project/cli/stringlib"
	"github.com/launchpad-project/cli/tdata"
	"github.com/launchpad-project/cli/verbose"
)

type postMock struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Comments int    `json:"comments"`
}

var bufErrStream bytes.Buffer

func TestMain(m *testing.M) {
	var defaultErrStream = errStream
	errStream = &bufErrStream

	globalconfigmock.Setup()
	ec := m.Run()
	globalconfigmock.Teardown()

	errStream = defaultErrStream

	os.Exit(ec)
}

func TestURLFailure(t *testing.T) {
	config.Stores["global"] = nil

	defer func() {
		r := recover()

		if r != errInvalidGlobalConfig {
			t.Errorf("Expected panic with %v error, got %v instead", errInvalidGlobalConfig, r)
		}

		globalconfigmock.Setup()
	}()

	URL("foo")
}

func TestAuth(t *testing.T) {
	r := launchpad.URL("http://localhost/")

	Auth(r)

	var want = "Basic YWRtaW46c2FmZQ==" // admin:safe in base64
	var got = r.Headers.Get("Authorization")

	if want != got {
		t.Errorf("Wrong auth header. Expected %s, got %s instead", want, got)
	}
}

func TestAuthGet(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/posts/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		fmt.Fprintf(w, `{
    "id": "1234",
    "title": "once upon a time",
    "body": "to be written",
    "comments": 30
}`)
	})

	var post postMock

	var wantID = "1234"
	var wantTitle = "once upon a time"
	var wantBody = "to be written"
	var wantComments = 30

	var err = AuthGet("/posts/1", &post)

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	if post.ID != wantID {
		t.Errorf("Wanted Id %v, got %v instead", wantID, post.ID)
	}

	if post.Title != wantTitle {
		t.Errorf("Wanted Title %v, got %v instead", wantTitle, post.Title)
	}

	if post.Body != wantBody {
		t.Errorf("Wanted Body %v, got %v instead", wantBody, post.Body)
	}

	if post.Comments != wantComments {
		t.Errorf("Wanted Comments %v, got %v instead", wantComments, post.Comments)
	}

	servertest.Teardown()
}

func TestAuthGetError(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	})

	var err = AuthGet("/foo", nil)

	switch err.(type) {
	case *APIFault:
		var ae = err.(*APIFault)

		if ae.Message != "Forbidden" {
			t.Errorf("Wanted Forbidden error message, got %v instead",
				ae.Message)
		}
	default:
		t.Errorf("Wrong error type, got %v instead", err)
	}

	servertest.Teardown()
}

func TestAuthGetOrExit(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/posts/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		fmt.Fprintf(w, `{
    "id": "1234",
    "title": "once upon a time",
    "body": "to be written",
    "comments": 30
}`)
	})

	var post postMock

	var wantID = "1234"
	var wantTitle = "once upon a time"
	var wantBody = "to be written"
	var wantComments = 30

	bufErrStream.Reset()

	AuthGetOrExit("/posts/1", &post)

	if post.ID != wantID {
		t.Errorf("Wanted Id %v, got %v instead", wantID, post.ID)
	}

	if post.Title != wantTitle {
		t.Errorf("Wanted Title %v, got %v instead", wantTitle, post.Title)
	}

	if post.Body != wantBody {
		t.Errorf("Wanted Body %v, got %v instead", wantBody, post.Body)
	}

	if post.Comments != wantComments {
		t.Errorf("Wanted Comments %v, got %v instead", wantComments, post.Comments)
	}

	if bufErrStream.Len() != 0 {
		t.Errorf("Unexpected content written to err stream: %v", bufErrStream.String())
	}

	servertest.Teardown()
}

func TestAuthGetOrExitError(t *testing.T) {
	servertest.Setup()
	haltExitCommand = true

	servertest.Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	})

	AuthGetOrExit("/foo", nil)

	var want = "WeDeploy API error: 403 Forbidden (GET http://www.example.com/foo)"

	if !strings.Contains(bufErrStream.String(), want) {
		t.Errorf("Missing expected error %v", want)
	}

	haltExitCommand = false
	servertest.Teardown()
}

func TestAuthTokenBearer(t *testing.T) {
	r := launchpad.URL("http://localhost/")

	var csg = config.Stores["global"]
	csg.Set("token", "mytoken")

	Auth(r)

	var want = "Bearer mytoken"
	var got = r.Headers.Get("Authorization")

	if want != got {
		t.Errorf("Wrong auth header. Expected %s, got %s instead", want, got)
	}
}

func TestAPIError(t *testing.T) {
	var e error = &APIFault{
		Code:    404,
		Message: "Resource Not Found",
	}

	if fmt.Sprintf("%v", e) != "WeDeploy API error: 404 Resource Not Found" {
		t.Errorf("APIFault error, got %v", e)
	}
}

func TestAPIFaultGet(t *testing.T) {
	var e = &APIFault{
		Code:    404,
		Message: "Resource Not Found",
		Errors: APIFaultErrors{
			APIFaultError{
				Reason:  "x",
				Message: "y",
			},
		},
	}

	var has, msg = e.Get("x")

	if !has {
		t.Errorf("Expected reason to exist.")
	}

	var want = "y"

	if msg != want {
		t.Errorf("Wanted reason to be %v, got %v", want, msg)
	}
}

func TestAPIFaultGetNotFound(t *testing.T) {
	var e = &APIFault{
		Code:    404,
		Message: "Resource Not Found",
		Errors:  APIFaultErrors{},
	}

	var has, msg = e.Get("x")

	if has || msg != "" {
		t.Errorf("Unexpected APIFault given error reason reported as existing")
	}
}

func TestAPIFaultGetNil(t *testing.T) {
	var e = &APIFault{
		Code:    404,
		Message: "Resource Not Found",
	}

	var has, msg = e.Get("x")

	if has || msg != "" {
		t.Errorf("Unexpected APIFault given error reason reported as existing")
	}
}

func TestAPIFaultHas(t *testing.T) {
	var e = &APIFault{
		Code:    404,
		Message: "Resource Not Found",
		Errors: APIFaultErrors{
			APIFaultError{
				Reason:  "x",
				Message: "y",
			},
		},
	}

	var has = e.Has("x")

	if !has {
		t.Errorf("Expected reason to exist.")
	}
}

func TestDecodeJSON(t *testing.T) {
	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/posts/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		fmt.Fprintf(w, `{
    "id": "1234",
    "title": "once upon a time",
    "body": "to be written",
    "comments": 30
}`)
	})

	var post postMock

	var wantID = "1234"
	var wantTitle = "once upon a time"
	var wantBody = "to be written"
	var wantComments = 30

	bufErrStream.Reset()

	r := URL("/posts/1")

	ValidateOrExit(r, r.Get())
	err := DecodeJSON(r, &post)

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	if post.ID != wantID {
		t.Errorf("Wanted Id %v, got %v instead", wantID, post.ID)
	}

	if post.Title != wantTitle {
		t.Errorf("Wanted Title %v, got %v instead", wantTitle, post.Title)
	}

	if post.Body != wantBody {
		t.Errorf("Wanted Body %v, got %v instead", wantBody, post.Body)
	}

	if post.Comments != wantComments {
		t.Errorf("Wanted Comments %v, got %v instead", wantComments, post.Comments)
	}

	if bufErrStream.Len() != 0 {
		t.Errorf("Unexpected content written to err stream: %v", bufErrStream.String())
	}
}

func TestDecodeJSONInvalidContentType(t *testing.T) {
	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/posts/1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
    "id": "1234",
    "title": "once upon a time",
    "body": "to be written",
    "comments": 30
}`)
	})

	var post postMock

	r := URL("/posts/1")

	ValidateOrExit(r, r.Get())
	err := DecodeJSON(r, &post)

	if err != ErrInvalidContentType {
		t.Errorf("Wanted error to be %v, got %v instead", ErrInvalidContentType, err)
	}
}

func TestDecodeJSONOrExit(t *testing.T) {
	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/posts/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		fmt.Fprintf(w, `{
    "id": "1234",
    "title": "once upon a time",
    "body": "to be written",
    "comments": 30
}`)
	})

	var post postMock

	var wantID = "1234"
	var wantTitle = "once upon a time"
	var wantBody = "to be written"
	var wantComments = 30

	haltExitCommand = true
	bufErrStream.Reset()

	r := URL("/posts/1")

	ValidateOrExit(r, r.Get())
	DecodeJSONOrExit(r, &post)

	if post.ID != wantID {
		t.Errorf("Wanted Id %v, got %v instead", wantID, post.ID)
	}

	if post.Title != wantTitle {
		t.Errorf("Wanted Title %v, got %v instead", wantTitle, post.Title)
	}

	if post.Body != wantBody {
		t.Errorf("Wanted Body %v, got %v instead", wantBody, post.Body)
	}

	if post.Comments != wantComments {
		t.Errorf("Wanted Comments %v, got %v instead", wantComments, post.Comments)
	}

	if bufErrStream.Len() != 0 {
		t.Errorf("Unexpected content written to err stream: %v", bufErrStream.String())
	}

	haltExitCommand = false
}

func TestDecodeJSONOrExitFailure(t *testing.T) {
	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/posts/1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		fmt.Fprintf(w, `[1234, 2010]`)
	})

	var post postMock

	var want = "json: cannot unmarshal array into Go value of type apihelper.postMock\n"

	haltExitCommand = true
	bufErrStream.Reset()

	r := URL("/posts/1/comments")

	ValidateOrExit(r, r.Get())
	DecodeJSONOrExit(r, &post)

	if bufErrStream.String() != want {
		t.Errorf("Wanted %v written to errStream, got %v instead", want, bufErrStream.String())
	}

	haltExitCommand = false
}

func TestEncodeJSON(t *testing.T) {
	type simple struct {
		Foo string `json:"foo"`
	}

	var m = &simple{
		Foo: "bar",
	}

	var foo, err = EncodeJSON(m)

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	var b bytes.Buffer
	foo.WriteTo(&b)

	var got = b.String()

	var want = `{"foo":"bar"}`

	if want != got {
		t.Errorf("Wanted encoded JSON to be %v, got %v instead", want, got)
	}
}

func TestEncodeJSONMap(t *testing.T) {
	var m = map[string]string{
		"foo": "bar",
	}

	var foo, err = EncodeJSON(m)

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	var b bytes.Buffer
	foo.WriteTo(&b)

	var got = b.String()

	var want = `{"foo":"bar"}`

	if want != got {
		t.Errorf("Wanted encoded JSON to be %v, got %v instead", want, got)
	}
}

func TestEncodeJSONMapUnsupportedType(t *testing.T) {
	var m = map[int]string{
		3: "bar",
	}

	var foo, err = EncodeJSON(m)

	var wantErr = "json: unsupported type: map[int]string"

	if err == nil || err.Error() != wantErr {
		t.Errorf("Wanted err to be %v, got %v instead", wantErr, wantErr)
	}

	var b bytes.Buffer
	foo.WriteTo(&b)

	var got = b.String()

	if len(got) != 0 {
		t.Errorf("Wanted unsupported JSON to have length 0, got %v instead", len(got))
	}
}

func TestParamsFromJSON(t *testing.T) {
	type musicianMock struct {
		ID       int64   `json:"id"`
		Name     string  `json:"name"`
		Hometown *string `json:"hometown"`
		LastName *string `json:"last_name,omitempty"`
		FullName *string `json:"full_name"`
		Friend   *string `json:"friend,omitempty"`
		Mic      bool    `json:"mic"`
		Age      int     `json:"age"`
		Height   float64 `json:"height"`
		password string
	}

	var fullName = "Ray Charles"
	var friend = "Gossie McKee"

	var musician = &musicianMock{
		ID:       14232,
		Name:     "Ray",
		FullName: &fullName,
		Friend:   &friend,
		Mic:      true,
		Age:      73,
		Height:   1.75,
		password: "c#swift",
	}

	var req = launchpad.URL("htt://example.com/")
	ParamsFromJSON(req, musician)

	var want = url.Values{
		"id":        []string{"14232"},
		"name":      []string{"Ray"},
		"full_name": []string{"Ray Charles"},
		"friend":    []string{"Gossie McKee"},
		"mic":       []string{"true"},
		"age":       []string{"73"},
		"height":    []string{"1.75"},
		"hometown":  []string{"null"},
	}

	var got = req.Params()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Params doesn't match:\n%s", pretty.Compare(want, got))
	}
}

func TestParamsFromJSONFailure(t *testing.T) {
	type invalidMock struct {
		AllowedParam      string   `json:"name"`
		NestingNotAllowed postMock `json:"post"`
	}

	var invalid = &invalidMock{}
	var req = launchpad.URL("htt://example.com/")

	defer func() {
		r := recover()

		if r != ErrExtractingParams {
			t.Errorf("Expected panic with %v error, got %v instead", ErrExtractingParams, r)
		}
	}()

	ParamsFromJSON(req, invalid)
}

func TestParamsFromJSONInvalidSstructure(t *testing.T) {
	var invalid = map[int]string{
		10: "foo",
	}
	var req = launchpad.URL("htt://example.com/")

	defer func() {
		r := recover()

		if r == nil {
			t.Errorf("Expected panic, got nil instead")
		}
	}()

	ParamsFromJSON(req, invalid)
}

func TestRequestVerboseFeedback(t *testing.T) {
	var defaultVerboseEnabled = verbose.Enabled
	var defaultVerboseErrStream = verbose.ErrStream
	var defaultNoColor = color.NoColor
	color.NoColor = true
	verbose.Enabled = true
	verbose.ErrStream = &bufErrStream
	bufErrStream.Reset()

	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Test-Multiple", "a")
		w.Header().Add("X-Test-Multiple", "b")
		fmt.Fprintf(w, "Hello")
	})

	var request = URL("/foo")

	request.Headers.Add("Accept", "application/json")
	request.Headers.Add("Accept", "text/plain")

	if err := request.Get(); err != nil {
		panic(err)
	}

	RequestVerboseFeedback(request)

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

	verbose.Enabled = defaultVerboseEnabled
	verbose.ErrStream = defaultVerboseErrStream
	color.NoColor = defaultNoColor
}

func TestRequestVerboseFeedbackUpload(t *testing.T) {
	var defaultVerboseEnabled = verbose.Enabled
	var defaultVerboseErrStream = verbose.ErrStream
	var defaultNoColor = color.NoColor
	color.NoColor = true
	verbose.Enabled = true
	verbose.ErrStream = &bufErrStream
	bufErrStream.Reset()

	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo", tdata.ServerHandler(""))

	var request = URL("/foo")

	var file, err = os.Open("mocks/config.json")

	if err != nil {
		panic(err)
	}

	request.Body(file)

	if err := request.Get(); err != nil {
		panic(err)
	}

	RequestVerboseFeedback(request)

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

	verbose.Enabled = defaultVerboseEnabled
	verbose.ErrStream = defaultVerboseErrStream
	color.NoColor = defaultNoColor
}

func TestRequestVerboseFeedbackStringReader(t *testing.T) {
	var defaultVerboseEnabled = verbose.Enabled
	var defaultVerboseErrStream = verbose.ErrStream
	var defaultNoColor = color.NoColor
	color.NoColor = true
	verbose.Enabled = true
	verbose.ErrStream = &bufErrStream
	bufErrStream.Reset()

	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo", tdata.ServerHandler(""))

	var request = URL("/foo")

	request.Body(strings.NewReader("custom body"))

	if err := request.Get(); err != nil {
		panic(err)
	}

	RequestVerboseFeedback(request)

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

	verbose.Enabled = defaultVerboseEnabled
	verbose.ErrStream = defaultVerboseErrStream
	color.NoColor = defaultNoColor
}

func TestRequestVerboseFeedbackBytesReader(t *testing.T) {
	var defaultVerboseEnabled = verbose.Enabled
	var defaultVerboseErrStream = verbose.ErrStream
	var defaultNoColor = color.NoColor
	color.NoColor = true
	verbose.Enabled = true
	verbose.ErrStream = &bufErrStream
	bufErrStream.Reset()

	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo", tdata.ServerHandler(""))

	var request = URL("/foo")

	var sr = strings.NewReader("custom body")

	var b bytes.Buffer
	sr.WriteTo(&b)

	request.Body(bytes.NewReader(b.Bytes()))

	if err := request.Get(); err != nil {
		panic(err)
	}

	RequestVerboseFeedback(request)

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

	verbose.Enabled = defaultVerboseEnabled
	verbose.ErrStream = defaultVerboseErrStream
	color.NoColor = defaultNoColor
}

func TestRequestVerboseFeedbackOtherReader(t *testing.T) {
	var defaultVerboseEnabled = verbose.Enabled
	var defaultVerboseErrStream = verbose.ErrStream
	var defaultNoColor = color.NoColor
	color.NoColor = true
	verbose.Enabled = true
	verbose.ErrStream = &bufErrStream
	bufErrStream.Reset()

	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo", tdata.ServerHandler(""))

	var request = URL("/foo")

	var body = strings.NewReader("custom body")

	var b = &bytes.Buffer{}

	request.Body(io.TeeReader(body, b))

	if err := request.Get(); err != nil {
		panic(err)
	}

	RequestVerboseFeedback(request)

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

	verbose.Enabled = defaultVerboseEnabled
	verbose.ErrStream = defaultVerboseErrStream
	color.NoColor = defaultNoColor
}

func TestRequestVerboseFeedbackJSONResponse(t *testing.T) {
	var defaultVerboseEnabled = verbose.Enabled
	var defaultVerboseErrStream = verbose.ErrStream
	var defaultNoColor = color.NoColor
	color.NoColor = true
	verbose.Enabled = true
	verbose.ErrStream = &bufErrStream
	bufErrStream.Reset()

	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"Hello": "World"}`)
	})

	var request = URL("/foo")

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

	RequestVerboseFeedback(request)

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

	verbose.Enabled = defaultVerboseEnabled
	verbose.ErrStream = defaultVerboseErrStream
	color.NoColor = defaultNoColor
}

func TestRequestVerboseFeedbackNullResponse(t *testing.T) {
	var defaultVerboseEnabled = verbose.Enabled
	var defaultVerboseErrStream = verbose.ErrStream
	var defaultNoColor = color.NoColor
	color.NoColor = true
	verbose.Enabled = true
	verbose.ErrStream = &bufErrStream
	bufErrStream.Reset()

	var request = URL("/foo")

	request.URL = "x://"

	// this returns an error, but we are not going to validate it here and
	// shortcut because we want to see what verbose returns
	request.Get()

	RequestVerboseFeedback(request)

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

	verbose.Enabled = defaultVerboseEnabled
	verbose.ErrStream = defaultVerboseErrStream
	color.NoColor = defaultNoColor
}

func TestRequestVerboseFeedbackNotComplete(t *testing.T) {
	var defaultVerboseEnabled = verbose.Enabled
	var defaultVerboseErrStream = verbose.ErrStream
	var defaultNoColor = color.NoColor
	color.NoColor = true
	verbose.Enabled = true
	verbose.ErrStream = &bufErrStream
	bufErrStream.Reset()

	RequestVerboseFeedback(URL("/foo"))

	stringlib.AssertSimilar(t,
		"> (wait) http://www.example.com/foo",
		bufErrStream.String())

	verbose.Enabled = defaultVerboseEnabled
	verbose.ErrStream = defaultVerboseErrStream
	color.NoColor = defaultNoColor
}

func TestSetBody(t *testing.T) {
	var got string
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		var body, err = ioutil.ReadAll(r.Body)

		if err != nil {
			t.Errorf("Wanted err to be nil, got %v instead", err)
		}

		got = string(body)
	})

	var request = URL("/foo")

	type simple struct {
		Foo string `json:"foo"`
	}

	var m = &simple{
		Foo: "bar",
	}

	var err = SetBody(request, m)

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	err = request.Get()

	if err != nil {
		t.Errorf("Wanted err to be nil, got %v instead", err)
	}

	var want = `{"foo":"bar"}`

	if want != got {
		t.Errorf("Wanted encoded JSON to be %v, got %v instead", want, got)
	}

	servertest.Teardown()
}

func TestSetBodyUnsupportedType(t *testing.T) {
	var request = URL("/foo")

	var m = map[int]string{
		3: "bar",
	}

	var err = SetBody(request, m)
	var wantErr = "json: unsupported type: map[int]string"

	if err == nil || err.Error() != wantErr {
		t.Errorf("Wanted err to be %v, got %v instead", wantErr, wantErr)
	}
}

func TestURL(t *testing.T) {
	var request = URL("x", "y", "z/k")
	var want = "http://www.example.com/x/y/z/k"

	if request.URL != want {
		t.Errorf("Wanted URL %v, got %v instead", want, request.URL)
	}
}

func TestValidate(t *testing.T) {
	var want = "Get x://localhost: unsupported protocol scheme \"x\""

	r := launchpad.URL("x://localhost/")

	err := Validate(r, r.Get())

	if err.Error() != want {
		t.Errorf("Wanted error to be %v, got %v instead", want, err)
	}
}

func TestValidateOrExit(t *testing.T) {
	var want = "Get x://localhost: unsupported protocol scheme \"x\"\n"
	haltExitCommand = true
	bufErrStream.Reset()

	r := launchpad.URL("x://localhost/")

	ValidateOrExit(r, r.Get())

	if bufErrStream.String() != want {
		t.Errorf("Wanted %v, got %v", want, bufErrStream.String())
	}

	haltExitCommand = false
}

func TestValidateOrExitNoError(t *testing.T) {
	haltExitCommand = true
	bufErrStream.Reset()

	r := launchpad.URL("x://localhost/")
	ValidateOrExit(r, nil)

	if bufErrStream.Len() != 0 {
		t.Errorf("Unexpected content written to err stream: %v", bufErrStream.String())
	}

	haltExitCommand = false
}

func TestValidateOrExitUnexpectedResponse(t *testing.T) {
	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo/bah", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		w.WriteHeader(403)
		fmt.Fprintf(w, `{
    "code": 403,
    "message": "Forbidden",
    "errors": [
        {
            "reason": "The requested operation failed because you do not have access.",
            "message": "forbidden"
        }
    ]
}`)
	})

	var want = "WeDeploy API error: 403 Forbidden (GET http://www.example.com/foo/bah)\n\t" +
		"forbidden: The requested operation failed because you do not have access.\n"

	haltExitCommand = true
	bufErrStream.Reset()

	r := URL("/foo/bah")

	ValidateOrExit(r, r.Get())

	if bufErrStream.String() != want {
		t.Errorf("Wanted %v, got %v", bufErrStream.String(), want)
	}

	haltExitCommand = false
}

func TestValidateOrExitUnexpectedNonJSONResponse(t *testing.T) {
	var defaultVerboseErrStream = verbose.ErrStream
	var defaultNoColor = color.NoColor
	color.NoColor = true
	verbose.Enabled = true
	verbose.ErrStream = &bufErrStream
	haltExitCommand = true
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo/bah", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json; charset=UTF-8")
		w.WriteHeader(403)
		fmt.Fprintf(w, `x`)
	})

	bufErrStream.Reset()

	var r = URL("/foo/bah")
	var err = Validate(r, r.Get())

	if err == nil {
		t.Errorf("Expected error, got %v instead", err)
	}

	var bes = bufErrStream.String()

	if !strings.Contains(bes,
		"Response not JSON (as expected by Content-Type)") {
		t.Errorf("Missing wrong response error")
	}

	if !strings.Contains(bes,
		"invalid character 'x' looking for beginning of value") {
		t.Errorf("Missing invalid error message")
	}

	color.NoColor = defaultNoColor
	verbose.Enabled = false
	verbose.ErrStream = defaultVerboseErrStream
	haltExitCommand = false
	servertest.Teardown()
}

func TestValidateOrExitUnexpectedResponseCustom(t *testing.T) {
	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo/bah", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		fmt.Fprintf(w, `Error message.`)
	})

	var want = tdata.FromFile("mocks/unexpected_response_error")
	haltExitCommand = true
	bufErrStream.Reset()

	r := URL("/foo/bah")

	ValidateOrExit(r, r.Get())

	if bufErrStream.String() != want {
		t.Errorf("Wanted %v, got %v", want, bufErrStream.String())
	}

	haltExitCommand = false
}
