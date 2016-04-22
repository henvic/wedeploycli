package tdata

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
)

type ResponseWriterMock struct {
	Want    string
	Test    *testing.T
	Headers http.Header
}

func (r *ResponseWriterMock) Header() http.Header {
	return r.Headers
}

func (r *ResponseWriterMock) Write(c []byte) (int, error) {
	var got = string(c)

	if r.Want != got {
		r.Test.Errorf("Wanted %v, got %v instead", r.Want, got)
	}

	return len(c), nil
}

func (*ResponseWriterMock) WriteHeader(status int) {}

func TestFromFile(t *testing.T) {
	var want = FromFile("mocks/mock")
	var got = "this is a mock\n"

	if want != got {
		t.Errorf("Wanted %v, got %v instead", got, want)
	}
}

func TestFromFileNotFound(t *testing.T) {
	var filename = fmt.Sprintf("not-found-%d", rand.Int())

	defer func() {
		r := recover()

		if !os.IsNotExist(r.(error)) {
			t.Errorf("Expected file %v to not exist", filename)
		}
	}()

	FromFile(filename)
}

func TestServerHandler(t *testing.T) {
	var handler = ServerHandler("this is a mock\n")
	var mock = &ResponseWriterMock{}
	mock.Want = "this is a mock\n"
	mock.Test = t
	handler(mock, nil)
}

func TestServerJSONHandler(t *testing.T) {
	var handler = ServerJSONHandler(`"this is a mock"`)
	var mock = &ResponseWriterMock{
		Headers: http.Header{},
	}
	mock.Want = `"this is a mock"`
	mock.Test = t
	handler(mock, nil)

	var want = "application/json; charset=UTF-8"
	var got = mock.Headers.Get("Content-Type")

	if got != want {
		t.Errorf("Wanted Content-Type %v, got %v instead", want, got)
	}
}

func TestServerFileHandler(t *testing.T) {
	var handler = ServerFileHandler("mocks/mock")
	var mock = &ResponseWriterMock{}
	mock.Want = "this is a mock\n"
	mock.Test = t
	handler(mock, nil)
}

func TestServerJSONFileHandler(t *testing.T) {
	var handler = ServerJSONFileHandler("mocks/mock.json")
	var mock = &ResponseWriterMock{
		Headers: http.Header{},
	}
	mock.Want = `"this is a mock"`
	mock.Test = t
	handler(mock, nil)

	var want = "application/json; charset=UTF-8"
	var got = mock.Headers.Get("Content-Type")

	if got != want {
		t.Errorf("Wanted Content-Type %v, got %v instead", want, got)
	}
}
