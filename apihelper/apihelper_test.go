package apihelper

import (
	"bytes"
	"context"
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

	"github.com/hashicorp/errwrap"
	wedeploy "github.com/henvic/wedeploy-sdk-go"
	"github.com/henvic/wedeploycli/color"
	"github.com/henvic/wedeploycli/config"
	"github.com/henvic/wedeploycli/defaults"
	"github.com/henvic/wedeploycli/servertest"
	"github.com/henvic/wedeploycli/stringlib"
	"github.com/henvic/wedeploycli/tdata"
	"github.com/henvic/wedeploycli/verbose"
	"github.com/kylelemons/godebug/pretty"
)

type postMock struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Body     string `json:"body"`
	Comments int    `json:"comments"`
}

var bufErrStream bytes.Buffer
var conf config.Context

func TestMain(m *testing.M) {
	var defaultErrStream = errStream
	errStream = &bufErrStream
	var err error
	conf, err = config.Setup("mocks/.lcp")

	if err != nil {
		panic(err)
	}

	if err := conf.SetEndpoint(defaults.CloudRemote); err != nil {
		panic(err)
	}

	ec := m.Run()

	errStream = defaultErrStream

	os.Exit(ec)
}

func TestAuth(t *testing.T) {
	r := wedeploy.URL("https://api.liferay.cloud/")

	(&Client{conf}).Auth(r)

	var want = "Bearer bar"
	var got = r.Headers.Get("Authorization")

	if want != got {
		t.Errorf("Wrong auth header. Expected %s, got %s instead", want, got)
	}
}

func TestAuthGet(t *testing.T) {
	servertest.Setup()

	servertest.Mux.HandleFunc("/posts/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, _ = fmt.Fprintf(w, `{
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

	var err = (&Client{conf}).AuthGet(context.Background(), "/posts/1", &post)

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

	var err = (&Client{conf}).AuthGet(context.Background(), "/foo", nil)

	switch ae := err.(type) {
	case APIFault:
		if ae.Message != "Forbidden" {
			t.Errorf("Wanted Forbidden error message, got %v instead",
				ae.Message)
		}
	default:
		t.Errorf("Wrong error type, got %v instead", err)
	}

	servertest.Teardown()
}

func TestAuthTokenBearer(t *testing.T) {
	r := wedeploy.URL("https://api.liferay.cloud/")

	(&Client{conf}).Auth(r)

	var want = "Bearer bar"
	var got = r.Headers.Get("Authorization")

	if want != got {
		t.Errorf("Wrong auth header. Expected %s, got %s instead", want, got)
	}
}

func TestAPIError(t *testing.T) {
	var e error = APIFault{
		Status:  404,
		Message: "Resource Not Found",
	}

	if fmt.Sprintf("%v", e) != "404 Not Found: Resource Not Found" {
		t.Errorf("APIFault error, got %v", e)
	}
}

func TestAPIFaultGet(t *testing.T) {
	var e = APIFault{
		Status:  404,
		Message: "Resource Not Found",
		Errors: APIFaultErrors{
			APIFaultError{
				Reason: "x",
				Context: APIFaultErrorContext{
					"message": "y",
				},
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

func TestAPIFaultUnstructured(t *testing.T) {
	var e = APIFault{
		URL:    "http://example.com/",
		Method: "POST",
		Status: 404,
	}

	var want = "POST http://example.com/ 404 Not Found"

	var msg = e.Error()

	if msg != want {
		t.Errorf("Wanted reason to be %v, got %v", want, msg)
	}
}

func TestAPIFaultMessageOnly(t *testing.T) {
	var e = APIFault{
		URL:     "http://example.com/",
		Method:  "POST",
		Status:  404,
		Message: "Resource Not Found",
	}

	var want = "POST http://example.com/ 404 Not Found: Resource Not Found"

	var msg = e.Error()

	if msg != want {
		t.Errorf("Wanted reason to be %v, got %v", want, msg)
	}
}

func TestAPIFaultContextMissingMessage(t *testing.T) {
	var e = APIFault{
		Status:  404,
		Message: "Resource Not Found",
		Errors: APIFaultErrors{
			APIFaultError{
				Reason: "x",
				Context: APIFaultErrorContext{
					"other": "abc",
				},
			},
		},
	}

	var want = "Reason (missing friendly message): x"
	var got = e.Error()

	if got != want {
		t.Errorf("Wanted reason to be %v, got %v", want, got)
	}
}

func TestAPIFaultGetNotFound(t *testing.T) {
	var e = APIFault{
		Status:  404,
		Message: "Resource Not Found",
		Errors:  APIFaultErrors{},
	}

	var has, msg = e.Get("x")

	if has || msg != "" {
		t.Errorf("Unexpected APIFault given error reason reported as existing")
	}
}

func TestAPIFaultGetNil(t *testing.T) {
	var e = APIFault{
		Status:  404,
		Message: "Resource Not Found",
	}

	var has, msg = e.Get("x")

	if has || msg != "" {
		t.Errorf("Unexpected APIFault given error reason reported as existing")
	}
}

func TestAPIFaultHas(t *testing.T) {
	var e = APIFault{
		Status:  404,
		Message: "Resource Not Found",
		Errors: APIFaultErrors{
			APIFaultError{
				Reason: "x",
				Context: APIFaultErrorContext{
					"message": "y",
				},
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
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, _ = fmt.Fprintf(w, `{
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

	r := (&Client{conf}).URL(context.Background(), "/posts/1")

	if err := Validate(r, r.Get()); err != nil {
		panic(err)
	}

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
		_, _ = fmt.Fprintf(w, `{
    "id": "1234",
    "title": "once upon a time",
    "body": "to be written",
    "comments": 30
}`)
	})

	var post postMock

	r := (&Client{conf}).URL(context.Background(), "/posts/1")

	if err := Validate(r, r.Get()); err != nil {
		panic(err)
	}

	err := DecodeJSON(r, &post)

	if err != ErrInvalidContentType {
		t.Errorf("Wanted error to be %v, got %v instead", ErrInvalidContentType, err)
	}
}

func TestDecodeJSONWithoutRequestResponse(t *testing.T) {
	var post postMock

	r := (&Client{conf}).URL(context.Background(), "/posts/1/comments")

	var err = DecodeJSON(r, &post)

	if err != errMissingResponse {
		t.Errorf("Wanted err to be %v, got %v instead", errMissingResponse, err)
	}
}

func TestDecodeJSONFailure(t *testing.T) {
	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/posts/1/comments", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		_, _ = fmt.Fprintf(w, `[1234, 2010]`)
	})

	var post postMock

	r := (&Client{conf}).URL(context.Background(), "/posts/1/comments")

	if err := Validate(r, r.Get()); err != nil {
		panic(err)
	}

	var err = DecodeJSON(r, &post)
	var wantErr = "json: cannot unmarshal array into Go value of type apihelper.postMock"
	var ew = errwrap.Contains(err, wantErr)

	if !ew {
		t.Errorf("Error message %v doesn't contain expected error %v", err, wantErr)
	}
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
	_, err = foo.WriteTo(&b)

	if err != nil {
		panic(err)
	}

	var got = b.String()

	var want = `{
    "foo": "bar"
}`

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
	_, err = foo.WriteTo(&b)

	if err != nil {
		panic(err)
	}

	var got = b.String()

	var want = `{
    "foo": "bar"
}`

	if want != got {
		t.Errorf("Wanted encoded JSON to be %v, got %v instead", want, got)
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

	var req = wedeploy.URL("htt://example.com/")
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
	var req = wedeploy.URL("htt://example.com/")

	defer func() {
		r := recover()

		if r != ErrExtractingParams {
			t.Errorf("Expected panic with %v error, got %v instead", ErrExtractingParams, r)
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
		_, _ = fmt.Fprintf(w, "Hello")
	})

	var request = (&Client{conf}).URL(context.Background(), "/foo")

	request.Headers.Add("Accept", "text/plain")

	if err := Validate(request, request.Get()); err != nil {
		panic(err)
	}

	var got = bufErrStream.String()

	var find = []string{
		"> GET https://api.liferay.cloud/foo HTTP/1.1",
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

	var request = (&Client{conf}).URL(context.Background(), "/foo")

	var file, err = os.Open("mocks/config.json")

	if err != nil {
		panic(err)
	}

	request.Body(file)

	if err := Validate(request, request.Get()); err != nil {
		panic(err)
	}

	var got = bufErrStream.String()

	var find = []string{
		"> GET https://api.liferay.cloud/foo HTTP/1.1",
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

	var request = (&Client{conf}).URL(context.Background(), "/foo")

	request.Body(strings.NewReader("custom body"))

	if err := Validate(request, request.Get()); err != nil {
		panic(err)
	}

	var got = bufErrStream.String()

	var find = []string{
		"> GET https://api.liferay.cloud/foo HTTP/1.1",
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

	var request = (&Client{conf}).URL(context.Background(), "/foo")

	var sr = strings.NewReader("custom body")

	var b bytes.Buffer
	_, err := sr.WriteTo(&b)

	if err != nil {
		panic(err)
	}

	request.Body(bytes.NewReader(b.Bytes()))

	if err := Validate(request, request.Get()); err != nil {
		panic(err)
	}

	var got = bufErrStream.String()

	var find = []string{
		"> GET https://api.liferay.cloud/foo HTTP/1.1",
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

	var request = (&Client{conf}).URL(context.Background(), "/foo")

	var body = strings.NewReader("custom body")

	var b = &bytes.Buffer{}

	request.Body(io.TeeReader(body, b))

	if err := Validate(request, request.Get()); err != nil {
		panic(err)
	}

	var got = bufErrStream.String()

	var find = []string{
		"> GET https://api.liferay.cloud/foo HTTP/1.1",
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
		_, _ = fmt.Fprintf(w, `{"Hello": "World"}`)
	})

	var request = (&Client{conf}).URL(context.Background(), "/foo")

	type Foo struct {
		Bar string `json:"bar"`
	}

	var foo = &Foo{Bar: "one"}

	var b, err = json.Marshal(foo)

	if err != nil {
		panic(err)
	}

	request.Body(bytes.NewBuffer(b))

	if err := Validate(request, request.Post()); err != nil {
		panic(err)
	}

	var got = bufErrStream.String()

	var find = []string{
		"> POST https://api.liferay.cloud/foo HTTP/1.1",
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

	var request = (&Client{conf}).URL(context.Background(), "/foo")

	request.URL = "x://"

	// this returns an error, but we are not going to shortcut to avoid getting
	// validation value here because we want to see what verbose returns
	var err = Validate(request, request.Get())

	if err == nil {
		t.Errorf("Expected error to be nil, got %v instead", err)
	}

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

	var request = (&Client{conf}).URL(context.Background(), "/foo")
	if err := Validate(request, nil); err != nil {
		panic(err)
	}

	stringlib.AssertSimilar(t,
		"> (wait) https://api.liferay.cloud/foo",
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

	var request = (&Client{conf}).URL(context.Background(), "/foo")

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

	var want = `{
    "foo": "bar"
}`

	if want != got {
		t.Errorf("Wanted encoded JSON to be %v, got %v instead", want, got)
	}

	servertest.Teardown()
}

func TestURL(t *testing.T) {
	var request = (&Client{conf}).URL(context.Background(), "x", "y", "z/k")
	var want = "https://api.liferay.cloud/x/y/z/k"

	if request.URL != want {
		t.Errorf("Wanted URL %v, got %v instead", want, request.URL)
	}
}

func TestValidate(t *testing.T) {
	r := wedeploy.URL("x://localhost/")
	err := Validate(r, r.Get())

	if want := "unsupported protocol scheme"; err == nil || !strings.Contains(err.Error(), want) {
		t.Errorf("Wanted error to be %v, got %v instead", want, err)
	}
}

func TestValidateNoError(t *testing.T) {
	r := wedeploy.URL("x://localhost/")

	if err := Validate(r, nil); err != nil {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestValidateUnexpectedResponse(t *testing.T) {
	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo/bah", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(403)
		_, _ = fmt.Fprintf(w, `{
    "status": 403,
    "message": "Forbidden",
    "errors": [
        {
            "reason": "The requested operation failed because you do not have access.",
			"context": {
				"message": "forbidden"
			}
        }
    ]
}`)
	})

	var want = "forbidden"

	r := (&Client{conf}).URL(context.Background(), "/foo/bah")
	err := Validate(r, r.Get())

	switch err.(type) {
	case APIFault:
	default:
		t.Errorf("Unexpected error type %v", reflect.TypeOf(err))
	}

	if err.Error() != want {
		t.Errorf("Wanted %v, got %v", want, err.Error())
	}
}

func TestValidateUnexpectedNonJSONResponse(t *testing.T) {
	var defaultVerboseErrStream = verbose.ErrStream
	var defaultNoColor = color.NoColor
	color.NoColor = true
	verbose.Enabled = true
	verbose.ErrStream = &bufErrStream
	servertest.Setup()

	servertest.Mux.HandleFunc("/foo/bah", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(403)
		_, _ = fmt.Fprintf(w, `x`)
	})

	bufErrStream.Reset()

	var r = (&Client{conf}).URL(context.Background(), "/foo/bah")
	var err = Validate(r, r.Get())

	if err == nil {
		t.Errorf("Expected error, got %v instead", err)
	}

	var bes = bufErrStream.String()

	if !strings.Contains(bes,
		"Invalid JSON response body:") {
		t.Errorf("Missing wrong response error")
	}

	if !strings.Contains(bes,
		"invalid character 'x' looking for beginning of value") {
		t.Errorf("Missing invalid error message")
	}

	color.NoColor = defaultNoColor
	verbose.Enabled = false
	verbose.ErrStream = defaultVerboseErrStream
	servertest.Teardown()
}

func TestValidateUnexpectedResponseCustom(t *testing.T) {
	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo/bah", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = fmt.Fprintf(w, `Error message.`)
	})

	var want = tdata.FromFile("mocks/unexpected_response_error")

	r := (&Client{conf}).URL(context.Background(), "/foo/bah")
	err := Validate(r, r.Get())

	if err == nil {
		t.Errorf("Expected error, got %v instead", err)
	}

	if err.Error() != want {
		t.Errorf("Wanted %v, got %v", want, err.Error())
	}
}

func TestValidateUnexpectedResponseNonBody(t *testing.T) {
	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo/bah", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	})

	var want = `403 Forbidden (GET https://api.liferay.cloud/foo/bah): Response Body is not JSON`

	r := (&Client{conf}).URL(context.Background(), "/foo/bah")
	err := Validate(r, r.Get())

	if err == nil {
		t.Errorf("Expected error, got nil instead")
	}

	if err.Error() != want {
		t.Errorf("Wanted %v, got %v", want, err.Error())
	}
}

func TestAPIFaultErrorContext(t *testing.T) {
	var c APIFaultErrorContext

	if c != nil {
		// err... pretty obvious it is...
		t.Errorf("Expected %v to be nil", c)
	}

	if c.Message() != "" {
		t.Errorf("Expected no error when testing for message of nil context")
	}
}
