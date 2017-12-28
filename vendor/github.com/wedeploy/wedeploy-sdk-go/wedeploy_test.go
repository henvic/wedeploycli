/*
Copyright (c) 2016-present, Liferay Inc. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice,
this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
this list of conditions and the following disclaimer in the documentation
and/or other materials provided with the distribution.

3. Neither the name of Liferay, Inc. nor the names of its contributors may
be used to endorse or promote products derived from this software without
specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/

package wedeploy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/wedeploy/wedeploy-sdk-go/aggregation"
	"github.com/wedeploy/wedeploy-sdk-go/filter"
	"github.com/wedeploy/wedeploy-sdk-go/jsonlib"
	"github.com/wedeploy/wedeploy-sdk-go/query"
)

var mux *http.ServeMux
var server *httptest.Server

func TestAuthBasic(t *testing.T) {
	r := URL("http://localhost/")
	r.Auth("admin", "safe")

	var want = "Basic YWRtaW46c2FmZQ==" // admin:safe in base64
	var got = r.Headers.Get("Authorization")

	if want != got {
		t.Errorf("Wrong auth header. Expected %s, got %s instead", want, got)
	}
}

func TestAuthBasicRequestParam(t *testing.T) {
	r := URL("http://localhost/")
	r.Auth("admin", "safe")

	err := r.setupAction("GET")

	if err != nil {
		t.Error(err)
	}

	var username, password, ok = r.Request.BasicAuth()

	if username != "admin" || password != "safe" || ok != true {
		t.Errorf("Wrong user credentials")
	}
}

func TestAuthOAuth(t *testing.T) {
	var want = "Bearer myToken"
	r := URL("http://localhost/")

	r.Auth("myToken")
	got := r.Headers.Get("Authorization")

	if want != got {
		t.Errorf("Wrong OAuth token. Wanted Bearer %s, got %s instead", want, got)
	}
}

func TestHeader(t *testing.T) {
	key := "X-Custom"
	value := "foo"
	req := URL("https://example.com/")
	req.Header(key, value)

	got := req.Headers.Get(key)

	if got != value {
		t.Errorf("Expected header %s=%s not found, got %s instead", key, value, got)
	}

}

func TestBodyRequest(t *testing.T) {
	setupServer()
	defer teardownServer()

	wantContentType := "text/plain"
	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `"body"`)

		gotContentType := r.Header.Get("Content-Type")
		assertTextualBody(t, "foo", r.Body)

		if gotContentType != wantContentType {
			t.Errorf("Expected content type %s, got %s instead",
				wantContentType,
				gotContentType)
		}
	})

	req := URL("http://example.com/url")

	req.Headers.Set("Content-Type", wantContentType)
	req.Body(bytes.NewBufferString("foo"))

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

func TestRequestBodyBuffered(t *testing.T) {
	// Buffered requested are removed by
	// *http.Client.Do
	// but we want wedeploy.RequestBody to "persist"
	// so we can read it afterwards (for example, for verbose mode)
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url")

	type Foo struct {
		Bar string `json:"bar"`
	}

	var foo = &Foo{Bar: "one"}
	var b, _ = json.Marshal(foo)

	req.Body(bytes.NewBuffer(b))

	if err := req.Get(); err != nil {
		t.Error(err)
	}

	var want = `{"bar":"one"}`
	var got = req.RequestBody.(*bytes.Buffer).String()

	if want != got {
		t.Errorf("Wanted request body %v, got %v instead", want, got)
	}
}

func TestRequestBodyBufferedWithRequestHeaderWithTimeout(t *testing.T) {
	// Buffered requested are removed by
	// *http.Client.Do
	// but we want wedeploy.RequestBody to "persist"
	// so we can read it afterwards (for example, for verbose mode)
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `"body"`)

		var want = "a-b-c"
		var got = r.Header.Get("Foo-Bar")

		if got != want {
			t.Errorf("Expected header value %v not found, found %v instead", want, got)
		}
	})

	req := URL("http://example.com/url")

	req.Header("Foo-Bar", "a-b-c")

	req.Timeout(10 * time.Second)

	type Foo struct {
		Bar string `json:"bar"`
	}

	var foo = &Foo{Bar: "one"}
	var b, _ = json.Marshal(foo)

	req.Body(bytes.NewBuffer(b))

	if err := req.Get(); err != nil {
		t.Error(err)
	}

	var want = `{"bar":"one"}`
	var got = req.RequestBody.(*bytes.Buffer).String()

	if want != got {
		t.Errorf("Wanted request body %v, got %v instead", want, got)
	}
}

func TestRequestBodyBufferedWithRequestHeaderWithTimeoutZero(t *testing.T) {
	// Buffered requested are removed by
	// *http.Client.Do
	// but we want wedeploy.RequestBody to "persist"
	// so we can read it afterwards (for example, for verbose mode)
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `"body"`)

		var want = "a-b-c"
		var got = r.Header.Get("Foo-Bar")

		if got != want {
			t.Errorf("Expected header value %v not found, found %v instead", want, got)
		}
	})

	req := URL("http://example.com/url")

	req.Header("Foo-Bar", "a-b-c")

	req.Timeout(0 * time.Second)

	type Foo struct {
		Bar string `json:"bar"`
	}

	var foo = &Foo{Bar: "one"}
	var b, _ = json.Marshal(foo)

	req.Body(bytes.NewBuffer(b))

	if err := req.Get(); err != nil {
		t.Error(err)
	}

	var want = `{"bar":"one"}`
	var got = req.RequestBody.(*bytes.Buffer).String()

	if want != got {
		t.Errorf("Wanted request body %v, got %v instead", want, got)
	}
}

func TestBodyRequestReadCloser(t *testing.T) {
	setupServer()
	defer teardownServer()

	var file, ferr = os.Open("LICENSE.md")

	if ferr != nil {
		panic(ferr)
	}

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		var a, er = ioutil.ReadAll(r.Body)
		if er != nil {
			t.Error(er)
		}

		var b, ef = ioutil.ReadFile("LICENSE.md")
		if ef != nil {
			t.Error(ef)
		}

		if bytes.Compare(a, b) != 0 {
			t.Errorf("Expected file received to be equal sent file")
		}
	})

	req := URL("http://example.com/url")

	req.Body(file)

	if err := req.Get(); err != nil {
		t.Error(err)
	}
}

func TestDecodeJSON(t *testing.T) {
	setupServer()
	defer teardownServer()

	var wantTitle = "body"

	setupDefaultMux(`{"title": "body"}`)

	req := URL("http://example.com/url")

	if err := req.Get(); err != nil {
		t.Error(err)
	}

	assertURI(t, "http://example.com/url", req.Request.URL.String())
	assertMethod(t, "GET", req.Request.Method)
	assertStatusCode(t, 200, req.Response.StatusCode)

	var content struct {
		Title string `json:"title"`
	}

	err := req.DecodeJSON(&content)

	if err != nil {
		t.Error(err)
	}

	if content.Title != wantTitle {
		t.Errorf("Expected title %s, got %s instead", wantTitle, content.Title)
	}
}

func TestDeleteRequest(t *testing.T) {
	setupServer()
	defer teardownServer()
	setupDefaultMux(`"body"`)

	req := URL("http://example.com/url")

	if err := req.Delete(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "DELETE", req.Request.Method)
}

func TestErrorStatusCode404(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	req := URL("http://example.com/url")

	if err := req.Get(); err != ErrUnexpectedResponse {
		t.Errorf("Missing error %s", ErrUnexpectedResponse)
	}

	assertTextualBody(t, "", req.Response.Body)
	assertStatusCode(t, 404, req.Response.StatusCode)
}

func TestGetRequest(t *testing.T) {
	setupServer()
	defer teardownServer()
	setupDefaultMux(`"body"`)

	req := URL("http://example.com/url")

	if err := req.Get(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "GET", req.Request.Method)
}

func TestGetRequestWithBackgroundContext(t *testing.T) {
	setupServer()
	defer teardownServer()
	setupDefaultMux(`"body"`)

	req := URL("http://example.com/url")
	req.SetContext(context.Background())

	if err := req.Get(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "GET", req.Request.Method)
}

func TestGetRequestWithBackgroundContextAndTimeout(t *testing.T) {
	setupServer()
	defer teardownServer()
	setupDefaultMux(`"body"`)

	req := URL("http://example.com/url")
	req.SetContext(context.Background())
	req.Timeout(time.Second)

	if err := req.Get(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "GET", req.Request.Method)
}

func TestGetRequestTimedout(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(800 * time.Millisecond)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url")
	req.Timeout(350 * time.Millisecond)

	switch err := req.Get(); err.(type) {
	case *url.Error:
		if !strings.Contains(err.Error(), "Get http://example.com/url: context deadline exceeded") {
			t.Errorf("Expected error due to client timeout, got %v instead", err)
		}
	default:
		t.Errorf("Expected error to be due to timeout, got %v instead", err)
	}
}

func TestGetRequestWithContextTimedout(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(800 * time.Millisecond)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url")
	req.SetContext(context.Background())
	req.Timeout(350 * time.Millisecond)

	switch err := req.Get(); err.(type) {
	case *url.Error:
		if !strings.Contains(err.Error(), "Get http://example.com/url: context deadline exceeded") {
			t.Errorf("Expected error due to client timeout, got %v instead", err)
		}
	default:
		t.Errorf("Expected error to be due to timeout, got %v instead", err)
	}
}

func TestGetRequestWithContextCanceledBeforeTimeout(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url")
	var cancelableCtx, cancel = context.WithCancel(context.TODO())
	req.SetContext(cancelableCtx)
	req.Timeout(350 * time.Millisecond)

	time.AfterFunc(100*time.Millisecond, func() {
		cancel()
	})

	switch err := req.Get(); err.(type) {
	case *url.Error:
		if !strings.Contains(err.Error(), "Get http://example.com/url: context canceled") {
			t.Errorf("Expected error due to client timeout, got %v instead", err)
		}
	default:
		t.Errorf("Expected error to be due to timeout, got %v instead", err)
	}
}

func TestHeadRequest(t *testing.T) {
	setupServer()
	defer teardownServer()

	wantHeader := "foo"

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Foo", wantHeader)
	})

	req := URL("http://example.com/url")

	if err := req.Head(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, "", req.Response.Body)
	assertMethod(t, "HEAD", req.Request.Method)

	gotHeader := req.Response.Header.Get("X-Foo")

	if wantHeader != gotHeader {
		t.Errorf("Want header X-Foo=%s, got %s instead", wantHeader, gotHeader)
	}
}

func TestPath(t *testing.T) {
	books := URL("https://example.com/books")
	book1 := books.Path("/1", "/2", "3")

	if books == book1 {
		t.Errorf("books and books1 should not be equal")
	}

	if books.URL != "https://example.com/books" {
		t.Error("books url is wrong")
	}

	want := "https://example.com/books/1/2/3"

	if book1.URL != want {
		t.Errorf("Unexpected book URL %s instead of %s", book1.URL, want)
	}
}

func TestUserAgent(t *testing.T) {
	r := URL("http://localhost/foo")
	err := r.setupAction("GET")

	if err != nil {
		t.Error(err)
	}

	var actual = r.Request.Header.Get("User-Agent")
	var expected = fmt.Sprintf("WeDeploy/%s (+https://wedeploy.com)", Version)

	if actual != expected {
		t.Errorf("Expected User-Agent %s doesn't match with %s", actual, expected)
	}
}

func TestURL(t *testing.T) {
	r := URL("https://example.com/foo/bah")

	if err := r.setupAction("GET"); err != nil {
		t.Error(err)
	}

	assertURI(t, "https://example.com/foo/bah", r.Request.URL.String())
}

func TestURLErrorDueToInvalidURI(t *testing.T) {
	setupServer()
	defer teardownServer()
	r := URL("://example.com/foo/bah")
	err := r.Get()
	kind := reflect.TypeOf(err).String()

	if kind != "*url.Error" {
		t.Errorf("Expected error *url.Error, got %s instead", kind)
	}
}

func TestParam(t *testing.T) {
	var req = URL("http://example.com/xyz?keep=this")

	req.Param("x", "i")
	req.Param("y", "j")
	req.Param("z", "k")

	var want = "http://example.com/xyz?keep=this&x=i&y=j&z=k"

	if req.URL != want {
		t.Errorf("Wanted url %v, got %v instead", want, req.URL)
	}
}

func TestParamOverwrite(t *testing.T) {
	var req = URL("http://example.com/xyz")

	req.Param("foo", "bar")
	req.Param("foo", "bar2")
	req.Param("foo", "bar3")

	var want = "http://example.com/xyz?foo=bar3"

	if req.URL != want {
		t.Errorf("Wanted url %v, got %v instead", want, req.URL)
	}
}

func TestParamParsingErrorSilentFailure(t *testing.T) {
	// Silently ignoring errors from parsing for now.
	// This test describes what happens when parsing errors exists.
	// Reason: API simplicity.

	// Any error triggered here should be triggered as soon as a REST action
	// such as Get() or Post() is called.
	// Never even worry about it. Never say never.

	// See also TestParamsParsingErrorSilentFailure

	var req = URL(":wrong-schema")

	req.Param("foo", "bar")

	var want = ":wrong-schema"

	if req.URL != want {
		t.Errorf("Wanted invalid url %v, got %v instead", want, req.URL)
	}
}

func TestParams(t *testing.T) {
	var want = url.Values{
		"q":    []string{"foo"},
		"page": []string{"2"},
	}

	var req = URL("http://google.com/?q=foo&page=2")
	var got = req.Params()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Params doesn't match:\n%s", pretty.Compare(want, got))
	}
}

func TestParamsParsingErrorSilentFailure(t *testing.T) {
	// See also TestParamParsingErrorSilentFailure
	var req = URL(":wrong-schema")
	var got = req.Params()

	if got != nil {
		t.Errorf("Params should be null, got %v instead", got)
	}
}

func TestPatchRequest(t *testing.T) {
	setupServer()
	defer teardownServer()
	setupDefaultMux(`"body"`)

	var req = URL("http://example.com/url")

	if err := req.Patch(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "PATCH", req.Request.Method)
}

func TestPostRequest(t *testing.T) {
	setupServer()
	defer teardownServer()
	setupDefaultMux(`"body"`)

	var req = URL("http://example.com/url")

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

func TestPostFormRequest(t *testing.T) {
	setupServer()
	defer teardownServer()

	wantContentType := "application/x-www-form-urlencoded"
	wantTitle := "foo"

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `"body"`)

		var gotContentType = r.Header.Get("Content-Type")
		var gotTitle = r.PostFormValue("title")

		if gotContentType != wantContentType {
			t.Errorf("Expected content type %s, got %s instead",
				wantContentType,
				gotContentType)
		}

		if gotTitle != wantTitle {
			t.Errorf("Expected title %s, got %s instead", wantTitle, gotTitle)
		}
	})

	req := URL("http://example.com/url")

	req.Form("title", wantTitle)

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

func TestPutRequest(t *testing.T) {
	setupServer()
	defer teardownServer()
	setupDefaultMux(`"body"`)

	var req = URL("http://example.com/url")

	if err := req.Put(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "PUT", req.Request.Method)
}

func TestQueryNull(t *testing.T) {
	req := URL("http://example.com/foo/bah")

	if req.Query != nil {
		t.Errorf("Expected empty query, found %v instead", req.Query)
	}
}

func TestQueryInvalidStructure(t *testing.T) {
	req := URL("http://example.com/foo/bah")
	type s struct{}
	var sv s

	req.Query = &query.Builder{
		BFilter: &[]filter.Filter{
			{"a": map[s]string{sv: "x"}},
		},
	}

	if err := req.setupAction("GET"); err == nil {
		t.Error("Expected Get() to fail due to invalid Query structure")
	}
}

func TestQueryString(t *testing.T) {
	setupServer()
	defer teardownServer()

	var want = "http://example.com/url?foo=bar"

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		var got = r.URL.String()
		if got != want {
			t.Errorf("Wanted URL %v, got %v instead", want, got)
		}
	})

	req := URL("http://example.com/url")

	req.Param("foo", "bar")
	req.Param("foo", "bar")

	if err := req.Get(); err != nil {
		t.Error(err)
	}
}

func TestQueryAggregate(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		assertBody(t, `{"aggregation":[{"bah":{"name":"foo"}}]}`, r.Body)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url")
	req.Aggregate("foo", "bah")

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

func TestQueryCount(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		assertBody(t, `{"type":"count"}`, r.Body)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url").Count()

	if req.Query.Type != "count" {
		t.Errorf("Expected count type, found %s instead", req.Query.Type)
	}

	if err := req.Get(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "GET", req.Request.Method)
}

func TestQueryFilter(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		assertBody(t, `{
    "filter": [
        {
            "foo": {
                "operator": "not",
                "value": "bah"
            }
        }
    ]
}`, r.Body)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url").Filter("foo", "not", "bah")

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

func TestQueryHighlight(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		assertBody(t, `{"highlight":["xyz"]}`, r.Body)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url").Highlight("xyz")

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

func TestQueryLimit(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		assertBody(t, `{"limit":10}`, r.Body)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url").Limit(10)

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

func TestQueryOffset(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		assertBody(t, `{"offset":0}`, r.Body)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url").Offset(0)

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

func TestQuerySort(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		assertBody(t, `{"sort":[{"id":"desc"}]}`, r.Body)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url").Sort("id", "desc")

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

func TestQuerySortAndAggregate(t *testing.T) {
	setupServer()
	defer teardownServer()

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		assertBody(t, `{
    "sort": [
        {
            "field2": "asc"
        }
    ],
    "aggregation": [
        {
            "f": {
                "operator": "min",
                "name": "a"
            }
        },
        {
            "f": {
                "operator": "missing",
                "name": "m"
            }
        }
    ]
}`, r.Body)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url")
	req.Sort("field2").Aggregate("a", "f", "min")
	req.Aggregate(aggregation.Missing("m", "f"))

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

func TestTimeout(t *testing.T) {
	var defaultTimeout = client.http.Timeout
	setupServer()
	defer teardownServer()
	defer func() {
		client.http.Timeout = defaultTimeout
	}()

	client.http.Timeout = 10 * time.Second

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintf(w, `"body"`)
	})

	req := URL("http://example.com/url")

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

func TestTimeoutFailure(t *testing.T) {
	var defaultTimeout = client.http.Timeout
	setupServer()
	defer teardownServer()
	defer func() {
		client.http.Timeout = defaultTimeout
	}()

	client.http.Timeout = 350 * time.Millisecond

	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second)
	})

	req := URL("http://example.com/url")

	switch err := req.Get(); err.(type) {
	case *url.Error:
		if !strings.Contains(err.Error(), "Client.Timeout") {
			t.Errorf("Expected error due to client timeout, got %v instead", err)
		}
	default:
		t.Errorf("Expected error to be due to timeout, got %v instead", err)
	}
}

func TestAlternative(t *testing.T) {
	setupServer()
	defer teardownServer()

	// change the client to force an error, if this fails
	Client().SetHTTP(&http.Client{})

	wantContentType := "text/plain"
	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `"body"`)

		gotContentType := r.Header.Get("Content-Type")
		assertTextualBody(t, "foo", r.Body)

		if gotContentType != wantContentType {
			t.Errorf("Expected content type %s, got %s instead",
				wantContentType,
				gotContentType)
		}
	})

	hc := NewHTTPClient()
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	hc.SetHTTP(&http.Client{
		Transport: transport,
	})

	req := hc.URL("http://example.com/url")

	req.Headers.Set("Content-Type", wantContentType)
	req.Body(bytes.NewBufferString("foo"))

	if err := req.Post(); err != nil {
		t.Error(err)
	}

	assertTextualBody(t, `"body"`, req.Response.Body)
	assertMethod(t, "POST", req.Request.Method)
}

// func TestConcurrency(t *testing.T) {
// 	go func() {
// 		Client = nil
// 	}()

// 	Client = nil
// }

func TestClient(t *testing.T) {
	if Client() != client {
		t.Errorf("Client auxiliary method is not returning client correctly")
	}
}

func assertBody(t *testing.T, want string, body io.ReadCloser) {
	bin, err := ioutil.ReadAll(body)

	if err != nil {
		t.Error(err)
	}

	got := make(map[string]interface{})
	err = json.Unmarshal(bin, &got)

	if err != nil {
		t.Error(err)
	}

	jsonlib.AssertJSONMarshal(t, want, got)
}

func assertStatusCode(t *testing.T, want int, got int) {
	if got != want {
		t.Errorf("Expected status code %d, got %d", want, got)
	}
}

func assertURI(t *testing.T, want, got string) {
	if got != want {
		t.Errorf("Expected URL %s, got %s", want, got)
	}
}

func assertMethod(t *testing.T, want, got string) {
	if got != want {
		t.Errorf("%s method expected, found %s instead", want, got)
	}
}

func assertTextualBody(t *testing.T, want string, got io.ReadCloser) {
	body, err := ioutil.ReadAll(got)

	if err != nil {
		t.Error(err)
	}

	var bString = string(body)

	if bString != want {
		t.Errorf("Expected body with %s, got %s instead", want, bString)
	}
}

func setupDefaultMux(content string) {
	mux.HandleFunc("/url", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, content)
	})
}

func setupServer() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	client.SetHTTP(&http.Client{Transport: transport})
}

func teardownServer() {
	client.SetHTTP(&http.Client{})
	server.Close()
}
