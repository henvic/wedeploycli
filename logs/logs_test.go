package logs

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/stringlib"
	"github.com/wedeploy/cli/tdata"
	"github.com/wedeploy/wedeploy-sdk-go/jsonlib"
)

var update bool

func init() {
	flag.BoolVar(&update, "update", false, "update golden files")
	setAnywhereOnEarthTimezone()
}

func setAnywhereOnEarthTimezone() {
	timezone := "Etc/GMT-12"
	l, err := time.LoadLocation(timezone)

	if err != nil {
		panic(err)
	}

	time.Local = l

	if err := os.Setenv("TZ", timezone); err != nil {
		panic(err)
	}
}

var wectx config.Context

func TestMain(m *testing.M) {
	var err error
	wectx, err = config.Setup("mocks/.lcp")

	if err != nil {
		panic(err)
	}

	if err := wectx.SetEndpoint(defaults.CloudRemote); err != nil {
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

	servertest.Mux.HandleFunc("/projects/foo/logs",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("serviceId") != "nodejs5143" {
				t.Errorf("Wrong value for serviceId")
			}

			if r.URL.Query().Get("level") != "4" {
				t.Errorf("Wrong value for level")
			}

			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			_, _ = fmt.Fprint(w, tdata.FromFile("mocks/logs_response.json"))
		})

	var filter = &Filter{
		Project:  "foo",
		Services: []string{"nodejs5143"},
		Instance: "foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj",
		Level:    4,
	}

	var list, err = New(wectx).GetList(context.Background(), filter)

	if err != nil {
		t.Errorf("Unexpected error %v on GetList", err)
	}

	jsonlib.AssertJSONMarshal(t, tdata.FromFile("mocks/logs_response_ref.json"), list)

	servertest.Teardown()
}

func TestList(t *testing.T) {
	outStreamMutex.Lock()
	var defaultOutStream = outStream
	outStream = &bufOutStream
	bufOutStream.Reset()
	outStreamMutex.Unlock()

	var defaultNoColor = color.NoColor
	color.NoColor = true

	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/logs",
		tdata.ServerJSONFileHandler("mocks/logs_response.json"))

	var filter = &Filter{
		Level:    4,
		Project:  "foo",
		Services: []string{"nodejs5143"},
		Instance: "foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj",
	}

	var err = New(wectx).List(context.Background(), filter)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	outStreamMutex.Lock()
	var got = bufOutStream.String()
	outStreamMutex.Unlock()

	if update {
		tdata.ToFile("mocks/logs_response_print", got)
	}

	var want = tdata.FromFile("mocks/logs_response_print")

	stringlib.AssertSimilar(t, want, got)

	color.NoColor = defaultNoColor
	outStreamMutex.Lock()
	outStream = defaultOutStream
	outStreamMutex.Unlock()

	servertest.Teardown()
}

func TestListProject(t *testing.T) {
	outStreamMutex.Lock()
	var defaultOutStream = outStream
	outStream = &bufOutStream
	bufOutStream.Reset()
	outStreamMutex.Unlock()

	var defaultNoColor = color.NoColor
	color.NoColor = true

	servertest.Setup()

	servertest.Mux.HandleFunc("/projects/foo/logs",
		tdata.ServerJSONFileHandler("mocks/logs_response_project.json"))

	var filter = &Filter{
		Level:   4,
		Project: "foo",
	}

	var err = New(wectx).List(context.Background(), filter)

	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	outStreamMutex.Lock()
	var got = bufOutStream.String()
	outStreamMutex.Unlock()

	if update {
		tdata.ToFile("mocks/logs_response_project_print", got)
	}

	var want = tdata.FromFile("mocks/logs_response_project_print")

	stringlib.AssertSimilar(t, want, got)

	color.NoColor = defaultNoColor
	outStreamMutex.Lock()
	outStream = defaultOutStream
	outStreamMutex.Unlock()

	servertest.Teardown()
}

func TestWatch(t *testing.T) {
	outStreamMutex.Lock()
	var defaultOutStream = outStream
	outStream = &bufOutStream
	bufOutStream.Reset()
	outStreamMutex.Unlock()

	var defaultNoColor = color.NoColor
	color.NoColor = true

	servertest.Setup()

	var missing = true

	servertest.Mux.HandleFunc("/projects/foo/logs",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("serviceId") != "bar" {
				t.Errorf("Wrong value for serviceId")
			}

			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			switch missing {
			case true:
				_, _ = fmt.Fprintln(w, tdata.FromFile("mocks/logs_watch_response_syscall.json"))
				missing = false
			default:
				_, _ = fmt.Fprintln(w, "[]")
			}
		})

	var w = &Watcher{
		Filter: &Filter{
			Project:  "foo",
			Services: []string{"bar"},
			Level:    4,
		},
		PoolingInterval: time.Millisecond,
	}

	var wg sync.WaitGroup

	wg.Add(1)

	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
		time.Sleep(20 * time.Millisecond)
		wg.Done()
	}()

	w.Watch(ctx, wectx)

	wg.Wait()

	outStreamMutex.Lock()
	var got = bufOutStream.String()
	outStreamMutex.Unlock()

	if update {
		tdata.ToFile("mocks/logs_watch_syscall", got)
	}

	var want = tdata.FromFile("mocks/logs_watch_syscall")

	stringlib.AssertSimilar(t, want, got)

	// some time before cleaning up services on other goroutines...
	time.Sleep(10 * time.Millisecond)
	color.NoColor = defaultNoColor
	outStreamMutex.Lock()
	outStream = defaultOutStream
	outStreamMutex.Unlock()
	servertest.Teardown()
}

func TestWatcherStart(t *testing.T) {
	outStreamMutex.Lock()
	var defaultOutStream = outStream
	outStream = &bufOutStream
	bufOutStream.Reset()
	outStreamMutex.Unlock()

	var defaultNoColor = color.NoColor
	color.NoColor = true

	servertest.Setup()

	var file = 0
	var fileM sync.Mutex

	servertest.Mux.HandleFunc("/projects/foo/logs",
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("serviceId") != "bar" {
				t.Errorf("Wrong value for serviceId")
			}

			fileM.Lock()
			defer fileM.Unlock()

			w.Header().Set("Content-Type", "application/json; charset=UTF-8")

			if file < 4 {
				file++
				log := fmt.Sprintf("%s%d%s", "mocks/logs_watch_response_", file, ".json")
				_, _ = fmt.Fprintln(w, tdata.FromFile(log))
			} else {
				_, _ = fmt.Fprintln(w, "[]")
			}
		})

	var w = &Watcher{
		Filter: &Filter{
			Project:  "foo",
			Services: []string{"bar"},
			Instance: "foo_nodejs5143_sqimupf5tfsf9iylzpg3e4zj",
			Level:    4,
		},
		PoolingInterval: 2 * time.Millisecond,
	}

	var wg sync.WaitGroup

	wg.Add(1)

	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// this sleep has to be slightly greater than pooling * requests
		time.Sleep(60 * time.Millisecond)
		cancel()
		wg.Done()
	}()

	w.Watch(ctx, wectx)

	wg.Wait()

	outStreamMutex.Lock()
	var got = bufOutStream.String()
	outStreamMutex.Unlock()

	if update {
		tdata.ToFile("mocks/logs_watch", got)
	}

	var want = tdata.FromFile("mocks/logs_watch")

	stringlib.AssertSimilar(t, want, got)

	// some time before cleaning up services on other goroutines...
	time.Sleep(10 * time.Millisecond)

	color.NoColor = defaultNoColor

	outStreamMutex.Lock()
	outStream = defaultOutStream
	outStreamMutex.Unlock()
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
