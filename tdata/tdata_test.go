package tdata

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
)

type ResponseWriterMock struct {
	test *testing.T
}

func (*ResponseWriterMock) Header() http.Header {
	return nil
}

func (r *ResponseWriterMock) SetTest(t *testing.T) {
	r.test = t
}
func (r *ResponseWriterMock) Write(c []byte) (int, error) {
	var want = "this is a mock\n"
	var got = string(c)

	if want != got {
		r.test.Errorf("Wanted %v, got %v instead", want, got)
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
	var handler = ServerHandler("mocks/mock")
	var mock = &ResponseWriterMock{}
	mock.SetTest(t)
	handler(mock, nil)
}
