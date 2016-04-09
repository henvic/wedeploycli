package apihelper

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/launchpad-project/api.go"
	"github.com/launchpad-project/cli/config"
	"github.com/launchpad-project/cli/globalconfigmock"
	"github.com/launchpad-project/cli/servertest"
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

func TestAuth(t *testing.T) {
	r := launchpad.URL("http://localhost/")

	Auth(r)

	var want = "Basic YWRtaW46c2FmZQ==" // admin:safe in base64
	var got = r.Headers.Get("Authorization")

	if want != got {
		t.Errorf("Wrong auth header. Expected %s, got %s instead", want, got)
	}
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

func TestDecodeJSON(t *testing.T) {
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

	var wantID = "1234"
	var wantTitle = "once upon a time"
	var wantBody = "to be written"
	var wantComments = 30

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

func TestDecodeJSONOrExit(t *testing.T) {
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

func TestParamsFromJSON(t *testing.T) {
	type musicianMock struct {
		ID       int64   `json:"id"`
		Name     string  `json:"name"`
		Hometown *string `json:"hometown"`
		LastName *string `json:"last_name,omitempty"`
		FullName *string `json:"full_name"`
		Friend   *string `json:"friend,omitempty"`
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
		t.Errorf("Wanted %v, got %v", bufErrStream.String(), want)
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
		w.WriteHeader(403)
		fmt.Fprintf(w, `{
    "code": 403,
    "message": "Forbidden",
    "errors": [
        {
            "reason": "forbidden",
            "message": "The requested operation failed because you do not have access."
        }
    ]
}`)
	})

	var want = "403 Forbidden\nThe requested operation failed because you do not have access.\n"
	haltExitCommand = true
	bufErrStream.Reset()

	r := URL("/foo/bah")

	ValidateOrExit(r, r.Get())

	if bufErrStream.String() != want {
		t.Errorf("Wanted %v, got %v", bufErrStream.String(), want)
	}

	haltExitCommand = false
}

func TestValidateOrExitUnexpectedResponseCustom(t *testing.T) {
	servertest.Setup()
	defer servertest.Teardown()

	servertest.Mux.HandleFunc("/foo/bah", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		fmt.Fprintf(w, `Error message.`)
	})

	var want = "403 Forbidden\nError message.\n"
	haltExitCommand = true
	bufErrStream.Reset()

	r := URL("/foo/bah")

	ValidateOrExit(r, r.Get())

	if bufErrStream.String() != want {
		t.Errorf("Wanted %v, got %v", bufErrStream.String(), want)
	}

	haltExitCommand = false
}
