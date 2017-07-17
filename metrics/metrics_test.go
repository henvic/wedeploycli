package metrics

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"encoding/json"

	"time"

	"reflect"

	"github.com/kylelemons/godebug/pretty"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/servertest"
	"github.com/wedeploy/cli/tdata"
)

func TestMain(m *testing.M) {
	removeMetricsFile()

	if err := config.Setup("mocks/.we"); err != nil {
		panic(err)
	}

	config.Global.EnableAnalytics = true
	resetMetricsSetup()
	ec := m.Run()
	SetPID(0)
	SetPath("")
	config.Teardown()
	os.Exit(ec)
}

func removeMetricsFile() {
	if err := os.Remove("mocks/.we_metrics"); err != nil && !os.IsNotExist(err) {
		panic(err)
	}
}

func resetMetricsSetup() {
	SetPID(os.Getpid())
	SetPath(abs("mocks/.we_metrics"))
	if err := Enable(); err != nil {
		panic(err)
	}
}

func TestConfigIsNil(t *testing.T) {
	var cached = config.Global
	defer func() {
		config.Global = cached
	}()

	config.Global = nil
	var saved, err = RecOrFail(Event{
		Type: "update",
		Text: "Starting update to version 0.2 from 0.1.",
		Tags: []string{
			"update_after_outdated_notification",
			"local_infrastructure_is_shutdown",
		},
		Extra: map[string]string{
			"old_version": "0.1",
			"new_version": "0.2",
			"channel":     "stable",
		},
	})

	if saved {
		t.Errorf("Expected config to not be saved")
	}

	if err != nil {
		t.Errorf("Expected no errors, got %v instead", err)
	}
}

func TestSetPath(t *testing.T) {
	defer resetMetricsSetup()

	var want = "foo/bar"
	SetPath(want)

	if metricsPath != want {
		t.Errorf("Wanted metrics path to be %v, got %v instead", want, metricsPath)
	}
}

func TestSetPID(t *testing.T) {
	defer resetMetricsSetup()

	var pid = SetPID(123)
	var sPid = strings.Split(pid, "-")

	if len(sPid) != 6 && sPid[len(sPid)-1] != "123" {
		t.Errorf("Wanted PID to be in format uuidv4-realPid, but it is not. Got %v instead", pid)
	}
}

func TestRecIsNotEnabled(t *testing.T) {
	config.Global.EnableAnalytics = false
	defer func() {
		config.Global.EnableAnalytics = true
	}()

	if config.Global.EnableAnalytics {
		t.Errorf("Analytics should not be enabled for this test")
	}

	saved, err := RecOrFail(Event{})

	if saved {
		t.Errorf("Event should not be saved: analytics should be disabled")
	}

	if err != nil {
		t.Errorf("Error should be nil, got %v instead", err)
	}
}

type testRecStruct struct {
	event  Event
	expect event
}

var testRecMock = []testRecStruct{
	testRecStruct{
		Event{
			Type: "update",
			Text: "Starting update to version 0.2 from 0.1.",
			Tags: []string{
				"update_after_outdated_notification",
				"local_infrastructure_is_shutdown",
			},
			Extra: map[string]string{
				"old_version": "0.1",
				"new_version": "0.2",
				"channel":     "stable",
			},
		},
		event{
			Type: "update",
			Text: "Starting update to version 0.2 from 0.1.",
			Tags: []string{
				"update_after_outdated_notification",
				"local_infrastructure_is_shutdown",
			},
			Extra: map[string]string{
				"channel":     "stable",
				"new_version": "0.2",
				"old_version": "0.1",
			},
			Scope:   "global",
			Version: defaults.Version,
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		},
	},
	testRecStruct{
		Event{
			Type: "install_docker_image",
			Text: "Installed docker image wedeploy/local:v1.0",
			Tags: []string{
				"local_infrastructure_is_updated",
			},
			Extra: map[string]string{
				"tag": "1.1",
			},
		},
		event{
			Type: "install_docker_image",
			Text: "Installed docker image wedeploy/local:v1.0",
			Tags: []string{"local_infrastructure_is_updated"},
			Extra: map[string]string{
				"tag": "1.1",
			},
			Scope:   "global",
			Version: defaults.Version,
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		},
	},
	testRecStruct{
		Event{
			Type: "we_run_ping",
			Text: "Server is up for 2 days",
			Tags: []string{
				"detached_mode",
			},
		},
		event{
			Type:    "we_run_ping",
			Text:    "Server is up for 2 days",
			Tags:    []string{"detached_mode"},
			Scope:   "global",
			Version: defaults.Version,
			OS:      runtime.GOOS,
			Arch:    runtime.GOARCH,
		},
	},
}

func TestRec(t *testing.T) {
	t.Run("testRec", testRec)
	t.Run("testReadingMetricsFile", testReadingMetricsFile)
	t.Run("testTrySubmitError", testTrySubmitError)
	t.Run("testTrySubmit", testTrySubmit)
}

func testRec(t *testing.T) {
	if _, err := os.Open("mocks/.we_metrics"); os.IsExist(err) {
		t.Fatalf(".we_metrics file should not exist at this time")
	}

	for _, e := range testRecMock {
		created, err := RecOrFail(e.event)

		if !created {
			t.Errorf("Should create event entry")
		}

		if err != nil {
			t.Errorf("Event error should be nil, got %v instead", err)
		}
	}
}

func testReadingMetricsFile(t *testing.T) {
	var lines = strings.Split(tdata.FromFile("mocks/.we_metrics"), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 lines, got %v instead", len(lines))
	}

	for n, e := range testRecMock {
		testRecCheckMetrics(t, lines[n], e.expect)
	}
}

func testTrySubmitError(t *testing.T) {
	var defaultEndpoint = defaults.AnalyticsEndpoint
	defer func() {
		defaults.AnalyticsEndpoint = defaultEndpoint
	}()

	var s = Sender{}
	defaults.AnalyticsEndpoint = "http://localhost:-1/"

	lines, err := s.TrySubmit()

	if err == nil || !strings.HasPrefix(err.Error(), "Can not submit analytics:") {
		t.Errorf("Expected error for TrySubmit() on invalid port not found, got %v instead", err)
	}

	if lines != 0 {
		t.Errorf("Lines should be zero")
	}
}

func testTrySubmit(t *testing.T) {
	var defaultEndpoint = defaults.AnalyticsEndpoint
	defer func() {
		defaults.AnalyticsEndpoint = defaultEndpoint
	}()

	servertest.SetupIntegration()
	defer servertest.TeardownIntegration()

	defaults.AnalyticsEndpoint = servertest.IntegrationServer.URL

	servertest.IntegrationMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var wantMethod = "POST"
		if r.Method != wantMethod {
			t.Errorf("Wanted method %v, got %v instead", wantMethod, r.Method)
		}

		var body, err = ioutil.ReadAll(r.Body)

		if err != nil {
			t.Errorf("Error parsing response")
		}

		var lines = strings.Split(string(body), "\n")

		if len(lines) != 4 {
			t.Errorf("Expected 4 lines, got %v instead", len(lines))
		}

		for n, e := range testRecMock {
			testRecCheckMetrics(t, lines[n], e.expect)
		}
	})

	t.Run("testTrySubmitWithoutPurging", testTrySubmitWithoutPurging)
	t.Run("testTrySubmitAndPurge", testTrySubmitAndPurge)
	t.Run("testTrySubmitNothingToSend", testTrySubmitNothingToSend)
}

func testTrySubmitWithoutPurging(t *testing.T) {
	var s = Sender{}
	lines, err := s.TrySubmit()

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if lines != 3 {
		t.Errorf("Lines should be 3, is %v instead", lines)
	}

	if _, err = os.Stat("mocks/.we_metrics"); err != nil {
		t.Errorf("we metrics file should exist, got %v error instead", err)
	}
}

func testTrySubmitAndPurge(t *testing.T) {
	var s = Sender{
		Purge: true,
	}

	lines, err := s.TrySubmit()

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if lines != 3 {
		t.Errorf("Lines should be 3, is %v instead", lines)
	}

	fi, ferr := os.Stat("mocks/.we_metrics")

	if ferr != nil {
		t.Errorf("we metrics file should exist, got %v error instead", ferr)
	}

	if fi.Size() != 0 {
		t.Errorf("Expected we metrics file to be truncated")
	}
}

func testTrySubmitNothingToSend(t *testing.T) {
	var s = Sender{
		Purge: true,
	}

	lines, err := s.TrySubmit()

	if err != nil {
		t.Errorf("Expected no error, got %v instead", err)
	}

	if lines != 0 {
		t.Errorf("Lines should be 0, is %v instead", lines)
	}
}

func testRecCheckMetrics(t *testing.T, line string, expected event) {
	var e event
	var err = json.Unmarshal([]byte(line), &e)

	if err != nil {
		t.Errorf("Expected no error unmarshalling JSON, got %v instead", err)
	}

	if len(e.ID) != 36 {
		t.Errorf("Expected ID to have 36 characters, got event.ID = %v instead", e.ID)
	}

	if len(e.PID) < 38 {
		t.Errorf("Expected ID to have at least 38 characters, got event.ID = %v instead", e.PID)
	}

	if len(e.SID) != 36 || e.SID != config.Global.AnalyticsID {
		t.Errorf("Expected SID to have 36 characters and be registered, got event.SID = %v) instead", e.SID)
	}

	if _, errt := time.Parse(time.RubyDate, e.Time); errt != nil {
		t.Errorf("Unexpected time parsing error for %v", e.Time)
	}

	// clearing ID, PID, SID, and Time to make it easier to compare values
	e.ID = ""
	e.PID = ""
	e.SID = ""
	e.Time = ""

	if !reflect.DeepEqual(e, expected) {
		t.Errorf("Expected event object doesn't match received value:\n%v\n",
			pretty.Compare(e, expected))
	}
}

func TestDisableAndEnableAndReset(t *testing.T) {
	var te = &testStatusMetricsStory{}
	t.Run("testDisable", te.testDisable)
	t.Run("testEnableAndRecording", te.testEnableAndRecording)
	t.Run("testResetting", te.testResetting)
}

type testStatusMetricsStory struct {
	initialSID string
}

func (te *testStatusMetricsStory) testDisable(t *testing.T) {
	te.initialSID = config.Global.AnalyticsID

	if err := Disable(); err != nil {
		t.Errorf("Expected no error while disabling analytics, got %v instead", err)
	}

	trySubmitCounter, trySubmitErr := (&Sender{}).TrySubmit()
	wantTrySubmitErr := "Aborting submission of analytics (analytics report status = disabled)"

	if trySubmitCounter != 0 ||
		trySubmitErr == nil ||
		trySubmitErr.Error() != wantTrySubmitErr {
		t.Errorf("Wanted TrySubmit error to be %v, got %v instead", wantTrySubmitErr, trySubmitErr)
	}

	if config.Global.EnableAnalytics {
		t.Errorf("Analytics should be disabled")
	}

	var weWithoutSpace = strings.Replace(tdata.FromFile("mocks/.we"), " ", "", -1)
	if !strings.Contains(weWithoutSpace, "enable_analytics=false") {
		t.Errorf(".we should have enable_analytics = false")
	}
}

func (te *testStatusMetricsStory) testEnableAndRecording(t *testing.T) {
	if err := Enable(); err != nil {
		t.Errorf("Expected no error while re-enabling analytics, got %v instead", err)
	}

	var weWithoutSpace = strings.Replace(tdata.FromFile("mocks/.we"), " ", "", -1)
	if !strings.Contains(weWithoutSpace, "enable_analytics=true") {
		t.Errorf(".we should have enable_analytics = true")
	}

	if config.Global.AnalyticsID != te.initialSID {
		t.Errorf("Initial SID should persist: wanted %v, got %v instead", te.initialSID, config.Global.AnalyticsID)
	}

	var bufErrStream bytes.Buffer
	var defaultErrStream = errStream
	errStream = &bufErrStream

	Rec(Event{
		Type: "test",
		Text: "Foo bar",
	})

	if len(tdata.FromFile("mocks/.we_metrics")) == 0 {
		t.Errorf(".we_metrics has no content")
	}

	if bufErrStream.Len() != 0 {
		t.Errorf("Expected no error in saving, got %v instead", bufErrStream.String())
	}

	errStream = defaultErrStream
}

func (te *testStatusMetricsStory) testResetting(t *testing.T) {
	if err := Reset(); err != nil {
		t.Errorf("Expected no error while re-enabling analytics, got %v instead", err)
	}

	if config.Global.AnalyticsID == te.initialSID {
		t.Errorf("Initial SID should be revoked: got same: %v", config.Global.AnalyticsID)
	}

	var weWithoutSpace = strings.Replace(tdata.FromFile("mocks/.we"), " ", "", -1)

	if strings.Contains(weWithoutSpace, "analytics_id="+te.initialSID) {
		t.Errorf(".we should not have analytics_id = <initial SID>")
	}

	if !strings.Contains(weWithoutSpace, "analytics_id="+config.Global.AnalyticsID) {
		t.Errorf(".we should have analytics_id = <new SID>")
	}

	if len(tdata.FromFile("mocks/.we_metrics")) != 0 {
		t.Errorf(".we_metrics should have no content")
	}
}

func TestPurgeError(t *testing.T) {
	var defaultMetricsPath = metricsPath
	defer func() {
		metricsPath = defaultMetricsPath
	}()

	metricsPath = os.DevNull

	if err := Purge(); err == nil {
		t.Errorf("Expected error when trying to purge unexisting metrics file")
	}
}

func TestSubmitMetricFileNotFound(t *testing.T) {
	var defaultMetricsPath = metricsPath
	defer func() {
		metricsPath = defaultMetricsPath
	}()

	metricsPath = abs("mocks/not-exists")

	var lines, err = (&Sender{}).TrySubmit()

	if err != nil {
		t.Errorf("TrySubmit error should be nil, got %v instead", err)
	}

	if lines != 0 {
		t.Errorf("Expected 0 lines, got %v instead", lines)
	}
}

func abs(path string) string {
	var abs, err = filepath.Abs(path)

	if err != nil {
		panic(err)
	}

	return abs
}
