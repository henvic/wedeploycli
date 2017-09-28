package logs

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/wedeploy/api-go/jsonlib"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/stringlib"
	"github.com/wedeploy/cli/tdata"
)

func TestMain(m *testing.M) {
	if err := config.Setup("mocks/.we"); err != nil {
		panic(err)
	}

	if err := config.SetEndpointContext(defaults.CloudRemote); err != nil {
		panic(err)
	}

	ec := m.Run()
	os.Exit(ec)
}

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

	servertest.Mux.HandleFunc("/projects/foo/services/nodejs5143/logs",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("level") != "4" {
				t.Errorf("Wrong value for level")
			}

			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			fmt.Fprintf(w, tdata.FromFile("mocks/logs_response.json"))
		})

	var filter = &Filter{
		Project:  "foo",
		Service:  "nodejs5143",
		Instance: "foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj",
		Level:    4,
	}

	var list, err = GetList(context.Background(), filter)

	if err != nil {
		t.Errorf("Unexpected error %v on GetList", err)
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile("mocks/logs_response_ref.json"), list)

	servertest.Teardown()
}

func TestList(t *testing.T) {
	var defaultOutStream = outStream
	outStream = &bufOutStream
	bufOutStream.Reset()

	var defaultNoColor = color.NoColor
	color.NoColor = true

	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/services/nodejs5143/logs",
		tdata.ServerJSONFileHandler("mocks/logs_response.json"))

	var filter = &Filter{
		Level:    4,
		Project:  "foo",
		Service:  "nodejs5143",
		Instance: "foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj",
	}

	var err = List(context.Background(), filter)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	var want = tdata.FromFile("mocks/logs_response_print")
	var got = bufOutStream.String()

	stringlib.AssertSimilar(t, want, got)

	color.NoColor = defaultNoColor
	outStream = defaultOutStream

	servertest.Teardown()
}

func TestListProject(t *testing.T) {
	var defaultOutStream = outStream
	outStream = &bufOutStream
	bufOutStream.Reset()

	var defaultNoColor = color.NoColor
	color.NoColor = true

	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/logs",
		tdata.ServerJSONFileHandler("mocks/logs_response_project.json"))

	var filter = &Filter{
		Level:   4,
		Project: "foo",
	}

	var err = List(context.Background(), filter)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	var want = tdata.FromFile("mocks/logs_response_project_print")
	var got = bufOutStream.String()

	stringlib.AssertSimilar(t, want, got)

	color.NoColor = defaultNoColor
	outStream = defaultOutStream

	servertest.Teardown()
}

func TestWatch(t *testing.T) {
	var defaultOutStream = outStream
	outStream = &bufOutStream
	bufOutStream.Reset()

	var defaultNoColor = color.NoColor
	color.NoColor = true

	servertest.Setup()

	var missing = true

	servertest.Mux.HandleFunc("/projects/foo/services/bar/logs",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			switch missing {
			case true:
				fmt.Fprintln(w, tdata.FromFile("mocks/logs_watch_response_syscall.json"))
				missing = false
			default:
				fmt.Fprintln(w, "[]")
			}
		})

	var watcher = &Watcher{
		Filter: &Filter{
			Project: "foo",
			Service: "bar",
			Level:   4,
		},
		PoolingInterval: time.Millisecond,
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		time.Sleep(20 * time.Millisecond)

		if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
			panic(err)
		}

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
	color.NoColor = defaultNoColor
	outStream = defaultOutStream
	servertest.Teardown()
}

func TestWatcherStart(t *testing.T) {
	var defaultOutStream = outStream
	outStream = &bufOutStream
	bufOutStream.Reset()

	var defaultNoColor = color.NoColor
	color.NoColor = true

	servertest.Setup()

	var fileNum = 0

	servertest.Mux.HandleFunc("/projects/foo/services/nodejs5143/logs",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
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
		Filter: &Filter{
			Project:  "foo",
			Service:  "nodejs5143",
			Instance: "foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj",
			Level:    4,
		},
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

	color.NoColor = defaultNoColor

	outStream = defaultOutStream
	servertest.Teardown()
}

func TestGetUnixTimestamp(t *testing.T) {
	var want int64 = 1470422556
	var got, err = GetUnixTimestamp("1470422556")

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	if want != got {
		t.Errorf("Expected numeric Unix timestamp return same numeric value")
	}
}

func TestGetUnixTimestampSame(t *testing.T) {
	var current = time.Now().Unix()
	var got, err = GetUnixTimestamp("60m60s")
	var diff int64 = 3660

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	// give some room for error, since it runs time.Now() again
	var delay = got - current + diff

	if delay < 0 || delay > 2 {
		t.Errorf("Wanted GetUnixTimestamp %v returned value not between expected boundaries", delay)
	}
}

func TestGetUnixTimestampSameMin(t *testing.T) {
	var current = time.Now().Unix()
	var got, err = GetUnixTimestamp("1h30min30s")
	var diff int64 = 5430

	if err != nil {
		t.Errorf("Wanted error to be nil, got %v instead", err)
	}

	// give some room for error, since it runs time.Now() again
	var delay = got - current + diff

	if delay < 0 || delay > 2 {
		t.Errorf("Wanted GetUnixTimestamp %v returned value not between expected boundaries", delay)
	}
}

func TestGetUnixTimestampParseError(t *testing.T) {
	var _, err = GetUnixTimestamp("dog")

	if err == nil {
		t.Errorf("Wanted parsing error, got %v instead", err)
	}
}
