package projects

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/launchpad-project/cli/globalconfigmock"
	"github.com/launchpad-project/cli/servertest"
)

var bufOutStream bytes.Buffer

func TestMain(m *testing.M) {
	var defaultOutStream = outStream
	outStream = &bufOutStream

	ec := m.Run()

	outStream = defaultOutStream
	os.Exit(ec)
}

func TestList(t *testing.T) {
	defer servertest.Teardown()
	servertest.Setup()
	globalconfigmock.Setup()
	bufOutStream.Reset()

	var want = "images (Image Server)\ntotal 1\n"

	servertest.Mux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `[{
    "name": "Image Server",
    "id": "images"
}]`)
	})

	List()

	if bufOutStream.String() != want {
		t.Errorf("Wanted %v, got %v instead", want, bufOutStream.String())
	}

	globalconfigmock.Teardown()
}
