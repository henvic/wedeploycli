package logs

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/launchpad-project/api.go/jsonlib"
	"github.com/launchpad-project/cli/globalconfigmock"
	"github.com/launchpad-project/cli/servertest"
	"github.com/launchpad-project/cli/stringlib"
	"github.com/launchpad-project/cli/tdata"
)

type GetLevelProvider struct {
	in    string
	out   int
	valid bool
}

var (
	bufOutStream bytes.Buffer
)

var GetLevelCases = []GetLevelProvider{
	{"0", 0, true},
	{"", 0, true},
	{"3", 3, true},
	{"critical", 2, true},
	{"error", 3, true},
	{"warning", 4, true},
	{"info", 6, true},
	{"debug", 7, true},
	{"foo", 0, false},
}

func TestGetLevel(t *testing.T) {
	for _, c := range GetLevelCases {
		out, err := GetLevel(c.in)
		valid := (c.valid == (err == nil))

		if out != c.out && valid {
			t.Errorf("Wanted level %v = (%v, valid: %v), got (%v, %v) instead",
				c.in,
				c.out,
				c.valid,
				out,
				err)
		}
	}
}

func TestGetList(t *testing.T) {
	servertest.Setup()
	globalconfigmock.Setup()

	servertest.Mux.HandleFunc("/api/logs/foo/nodejs5143/foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj",
		tdata.ServerFileHandler("mocks/logs_response.json"))

	var filter = &Filter{
		Level: 4,
	}

	var args = []string{"foo", "nodejs5143", "foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj"}

	var list = GetList(filter, args...)

	jsonlib.AssertJSONMarshal(t, tdata.FromFile("mocks/logs_response_ref.json"), list)

	globalconfigmock.Teardown()
	servertest.Teardown()
}

func TestList(t *testing.T) {
	var defaultOutStream = outStream
	outStream = &bufOutStream
	bufOutStream.Reset()

	globalconfigmock.Setup()
	servertest.Setup()

	servertest.Mux.HandleFunc("/api/logs/foo/nodejs5143/foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj",
		tdata.ServerFileHandler("mocks/logs_response.json"))

	var filter = &Filter{
		Level: 4,
	}

	var args = []string{"foo", "nodejs5143", "foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj"}

	List(filter, args...)

	var want = tdata.FromFile("mocks/logs_response_print")
	var got = bufOutStream.String()

	stringlib.AssertSimilar(t, want, got)

	outStream = defaultOutStream

	globalconfigmock.Teardown()
	servertest.Teardown()
}

func TestWatch(t *testing.T) {
	var defaultOutStream = outStream
	outStream = &bufOutStream
	bufOutStream.Reset()

	globalconfigmock.Setup()
	servertest.Setup()

	var missing = true

	servertest.Mux.HandleFunc("/api/logs/foo/bar",
		func(w http.ResponseWriter, r *http.Request) {
			switch missing {
			case true:
				fmt.Fprintln(w, tdata.FromFile("mocks/logs_watch_response_syscall.json"))
				missing = false
			default:
				fmt.Fprintln(w, "[]")
			}
		})

	var watcher = &Watcher{
		Filter: &Filter{Level: 4},
		Paths: []string{
			"foo",
			"bar"},
		PoolingInterval: time.Millisecond,
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		time.Sleep(20 * time.Millisecond)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		time.Sleep(20 * time.Millisecond)
		wg.Done()
	}()

	Watch(watcher)

	wg.Wait()

	var want = tdata.FromFile("mocks/logs_watch_syscall")
	var got = bufOutStream.String()

	stringlib.AssertSimilar(t, want, got)

	// some time before cleaning up services on other goroutines...
	time.Sleep(10 * time.Millisecond)
	outStream = defaultOutStream
	globalconfigmock.Teardown()
	servertest.Teardown()
}

func TestWatcherStart(t *testing.T) {
	var defaultOutStream = outStream
	outStream = &bufOutStream
	bufOutStream.Reset()
	servertest.Setup()
	globalconfigmock.Setup()

	var fileNum = 0

	servertest.Mux.HandleFunc("/api/logs/foo/nodejs5143/foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj",
		func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * time.Millisecond)
			if fileNum < 4 {
				fileNum++
				log := fmt.Sprintf("%s%d%s", "mocks/logs_watch_response_", fileNum, ".json")
				fmt.Fprintln(w, tdata.FromFile(log))
			} else {
				fmt.Fprintln(w, "[]")
			}
		})

	var watcher = &Watcher{
		Filter: &Filter{Level: 4},
		Paths: []string{
			"foo",
			"nodejs5143",
			"foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj"},
		PoolingInterval: time.Millisecond,
	}

	done := make(chan bool, 1)

	go func() {
		watcher.Start()
		// this sleep has to be slightly greater than pooling * requests
		time.Sleep(60 * time.Millisecond)
		watcher.Stop()
		done <- true
	}()

	<-done

	var want = tdata.FromFile("mocks/logs_watch")
	var got = bufOutStream.String()

	stringlib.AssertSimilar(t, want, got)

	// some time before cleaning up services on other goroutines...
	time.Sleep(10 * time.Millisecond)
	outStream = defaultOutStream
	servertest.Teardown()
	globalconfigmock.Teardown()
}
