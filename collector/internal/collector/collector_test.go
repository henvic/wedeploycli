package collector

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestHandlingMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	wantCode := http.StatusMethodNotAllowed

	if w.Code != wantCode {
		t.Errorf("Wanted status code to be %v, got %v instead", wantCode, w.Code)
	}

	wantBody := "Method Not Allowed: metrics collector only accepts POST requests\n"

	if w.Body.String() != wantBody {
		t.Errorf("Expected body to be %v, got %v instead", wantBody, w.Body.String())
	}
}

func TestHandlingEmpty(t *testing.T) {
	req := httptest.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	wantCode := http.StatusOK

	if w.Code != wantCode {
		t.Errorf("Wanted status code to be %v, got %v instead", wantCode, w.Code)
	}

	var want = BulkCollectorFeedback{
		Errors:           false,
		Objects:          0,
		JSONFailureLines: []int{},
		Insertions:       nil,
	}

	jsonlib.AssertJSONMarshal(t, w.Body.String(), want)
}

func TestHandlingBackendUnavailable(t *testing.T) {
	var b = bytes.Buffer{}
	log.SetOutput(&b)

	defer func() {
		log.SetOutput(os.Stderr)
	}()

	body := &bytes.Buffer{}
	_, _ = body.WriteString("{}\n")
	req := httptest.NewRequest("POST", "/", body)
	w := httptest.NewRecorder()
	Handler(w, req)

	wantCode := http.StatusServiceUnavailable

	if w.Code != wantCode {
		t.Errorf("Wanted status code to be %v, got %v instead", wantCode, w.Code)
	}

	var wantLogErr = `Error trying to POST to backend: Post events/_bulk: unsupported protocol scheme ""`
	if !strings.Contains(b.String(), wantLogErr) {
		t.Errorf("Expected error log to contain %v, but not found on %v", wantLogErr, b.String())
	}

	var wantErrFeedback = "Error trying to store analytics."

	if w.Body.String() != wantErrFeedback {
		t.Errorf("Wanted feedback to be %v, got %v instead", wantErrFeedback, w.Body.String())
	}
}

func TestHandling(t *testing.T) {
	servertest.SetupIntegration()
	Backend = servertest.IntegrationServer.URL
	var b = bytes.Buffer{}
	log.SetOutput(&b)

	defer func() {
		servertest.TeardownIntegration()
		Backend = ""
		log.SetOutput(os.Stderr)
	}()

	var unavailableTest = false

	servertest.IntegrationMux.HandleFunc("/events/_bulk",
		func(w http.ResponseWriter, r *http.Request) {
			if !unavailableTest {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w, "Unavailable")
				unavailableTest = true
				return
			}

			if r.Method != http.MethodPost {
				t.Errorf("Wanted method to be %v, got %v instead", http.MethodPost, r.Method)
			}

			var bBody, err = ioutil.ReadAll(r.Body)

			if err != nil {
				t.Errorf("Got %v error, wanted nil instead", err)
			}

			if !strings.Contains(string(bBody), `{"create": {"_index": "we-cli", "_type": "metrics", "_id": "foo"}}`) {
				t.Errorf("Expected string to contain bulk header for first item")
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/backend_bulk_response.json"))
		})

	body := &bytes.Buffer{}
	_, _ = body.WriteString(`{"id": "foo", "text": "example"}

{"id": "bar", "text": "example2"}`)
	req := httptest.NewRequest("POST", "/", body)
	w := httptest.NewRecorder()
	Handler(w, req)

	wantCode := http.StatusOK

	if w.Code != wantCode {
		t.Errorf("Wanted status code to be %v, got %v instead", wantCode, w.Code)
	}

	if !unavailableTest {
		t.Errorf("Expected unavailable test flag to be true")
	}
}

func TestHandlingBackendObjectError(t *testing.T) {
	servertest.SetupIntegration()
	Backend = servertest.IntegrationServer.URL
	var b = bytes.Buffer{}
	log.SetOutput(&b)

	defer func() {
		servertest.TeardownIntegration()
		Backend = ""
		log.SetOutput(os.Stderr)
	}()

	var unavailableTest = false

	servertest.IntegrationMux.HandleFunc("/events/_bulk",
		func(w http.ResponseWriter, r *http.Request) {
			if !unavailableTest {
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w, "Unavailable")
				unavailableTest = true
				return
			}

			if r.Method != http.MethodPost {
				t.Errorf("Wanted method to be %v, got %v instead", http.MethodPost, r.Method)
			}

			var body, err = ioutil.ReadAll(r.Body)

			if err != nil {
				panic(err)
			}

			if !strings.Contains(string(body), `{"create": {"_index": "we-cli", "_type": "metrics", "_id": "foo"}}`) {
				t.Errorf("Expected string to contain bulk header for first item")
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/backend_bulk_response.json"))
		})

	body := &bytes.Buffer{}
	_, _ = body.WriteString(`{"id": "foo", "text": "example"}
xyz
{"id": "bar", "text": "example2"}`)
	req := httptest.NewRequest("POST", "/", body)
	w := httptest.NewRecorder()
	Handler(w, req)

	wantCode := http.StatusOK

	if w.Code != wantCode {
		t.Errorf("Wanted status code to be %v, got %v instead", wantCode, w.Code)
	}

	if !unavailableTest {
		t.Errorf("Expected unavailable test flag to be true")
	}

	var want = BulkCollectorFeedback{
		Errors:           false,
		Objects:          3,
		JSONFailureLines: []int{2, 2},
		Insertions: []Insertion{
			Insertion{
				ID:     "aasxdasdabc",
				Error:  false,
				Line:   0,
				Status: 201,
			},
			Insertion{
				ID:     "axxbcd",
				Error:  false,
				Line:   0,
				Status: 201,
			},
		},
	}
	jsonlib.AssertJSONMarshal(t, w.Body.String(), want)
}

func TestHandlingMultiple(t *testing.T) {
	servertest.SetupIntegration()
	Backend = servertest.IntegrationServer.URL
	var b = bytes.Buffer{}
	log.SetOutput(&b)

	defer func() {
		servertest.TeardownIntegration()
		Backend = ""
		log.SetOutput(os.Stderr)
	}()

	var try = 0

	servertest.IntegrationMux.HandleFunc("/events/_bulk",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Wanted method to be %v, got %v instead", http.MethodPost, r.Method)
			}

			var _, err = ioutil.ReadAll(r.Body)

			if err != nil {
				t.Fatalf("Expected no error on getting body, got %v instead", err)
			}

			w.Header().Set("Content-type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/backend_bulk_response.json"))
			try++
		})

	body := &bytes.Buffer{}
	for counter := 1; counter <= 130; counter++ {
		_, _ = body.WriteString(fmt.Sprintf("{\"id\": \"foo-%v\", \"text\": \"example\"}\n", counter))
	}

	req := httptest.NewRequest("POST", "/", body)
	w := httptest.NewRecorder()
	Handler(w, req)

	wantCode := http.StatusOK

	if w.Code != wantCode {
		t.Errorf("Wanted status code to be %v, got %v instead", wantCode, w.Code)
	}

	if try != 3 {
		t.Errorf("Expected 3 requests to the backend server, got %v instead", try)
	}
}
