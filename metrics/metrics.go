package metrics

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"runtime"
	"time"

	"strings"

	"encoding/json"

	"os/exec"

	"github.com/hashicorp/errwrap"
	uuid "github.com/satori/go.uuid"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/defaults"
	"github.com/wedeploy/cli/verbose"
	"github.com/wedeploy/cli/verbosereq"
	wedeploy "github.com/wedeploy/wedeploy-sdk-go"
)

var (
	pid         string
	metricsPath string
	server      = defaults.AnalyticsEndpoint

	errStream io.Writer = os.Stderr
)

const (
	// MetricsSubmissionTimeout is the timeout for a metrics submission request
	MetricsSubmissionTimeout = 45 * time.Second
)

// SetPID enhances real PID with prefix to distinguish between repeating PIDs
func SetPID(p int) (uniquePID string) {
	// Enhance real PID with prefix to distinguish between repeating PIDs
	pid = fmt.Sprintf("%v-%v", newAnalyticsID(), p)
	return pid
}

// SetPath sets the path for the metric file
func SetPath(path string) {
	metricsPath = path
}

// Event entry
// see also collector.Event
type Event struct {
	Type  string
	Text  string
	Tags  []string
	Extra map[string]string
}

// internal event object
// used to separate "reserved fields" only
type event struct {
	ID      string            `json:"id"`
	Type    string            `json:"event_type,omitempty"`
	Text    string            `json:"text,omitempty"`
	Tags    []string          `json:"tags,omitempty"`
	Extra   map[string]string `json:"extra,omitempty"`
	PID     string            `json:"pid,omitempty"`
	SID     string            `json:"sid,omitempty"`
	Time    string            `json:"time,omitempty"`
	Version string            `json:"version,omitempty"`
	OS      string            `json:"os,omitempty"`
	Arch    string            `json:"arch,omitempty"`
}

// Rec event if analytics is enabled
func Rec(conf *config.Config, e Event) {
	if _, err := RecOrFail(conf, e); err != nil {
		fmt.Fprintf(errStream, "%v\n", err)
	}
}

// RecOrFail records event if analytics is enabled and returns an error on failure
func RecOrFail(conf *config.Config, e Event) (saved bool, err error) {
	if conf == nil || !conf.EnableAnalytics {
		return false, nil
	}

	var ie = event{
		ID:      newAnalyticsID(),
		Type:    e.Type,
		Text:    e.Text,
		Tags:    e.Tags,
		Extra:   e.Extra,
		PID:     pid,
		SID:     conf.AnalyticsID,
		Time:    time.Now().Format(time.RubyDate),
		Version: defaults.Version,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}

	if err = (&eventRecorder{}).rec(ie); err != nil {
		return false, errwrap.Wrapf("Error trying to record analytics report: {{err}}", err)
	}

	return true, nil
}

// Enable metrics
func Enable(conf *config.Config) error {
	conf.EnableAnalytics = true

	if conf.AnalyticsID == "" {
		conf.AnalyticsID = newAnalyticsID()
	}

	return conf.Save()
}

// Disable metrics
func Disable(conf *config.Config) error {
	conf.EnableAnalytics = false
	return conf.Save()
}

// Reset metrics by regenerating metrics ID and purge existing metrics
func Reset(conf *config.Config) error {
	conf.AnalyticsID = newAnalyticsID()

	if err := conf.Save(); err != nil {
		return err
	}

	return Purge()
}

func newAnalyticsID() string {
	return uuid.NewV4().String()
}

type eventRecorder struct {
	event         event
	jsonMarshaled []byte
}

func (e *eventRecorder) jsonMarshal() (err error) {
	if e.jsonMarshaled, err = json.Marshal(e.event); err != nil {
		return errwrap.Wrapf("Can not JSON marshal metrics event: {{err}}", err)
	}

	return nil
}

func (e *eventRecorder) appendEvent() error {
	var file, err = os.OpenFile(metricsPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer func() {
		if ef := file.Close(); ef != nil {
			verbose.Debug("Error appending metric to metrics file: %v", ef)
		}
	}()

	if err != nil {
		return errwrap.Wrapf("Can not open metrics file: {{err}}", err)
	}

	if _, err = file.Write(append(e.jsonMarshaled, []byte("\n")...)); err != nil {
		return errwrap.Wrapf("Error trying to write to metrics file: {{err}}", err)
	}

	return nil
}

func (e *eventRecorder) rec(ie event) (err error) {
	e.event = ie
	if err = e.jsonMarshal(); err != nil {
		return err
	}

	return e.appendEvent()
}

// TrySubmit events file if enabled
func (s *Sender) TrySubmit(conf *config.Config) (int, error) {
	if !conf.EnableAnalytics {
		return 0, errors.New(
			"Aborting submission of analytics (analytics report status = disabled)")
	}

	verbose.Debug("Submitting analytics")
	var lines, err = s.trySend()

	if err != nil {
		return 0, errwrap.Wrapf("Can not submit analytics: {{err}}", err)
	}

	if s.Purge {
		err = Purge()
	}

	if err != nil {
		return lines, errwrap.Wrapf("Error trying to purge analytics file: {{err}}", err)
	}

	return lines, nil
}

// Sender for the metrics
type Sender struct {
	Purge   bool
	content []byte
	lines   int
}

func (s *Sender) trySend() (events int, err error) {
	if err = s.read(); err != nil {
		return 0, err
	}

	if s.lines == 0 {
		return 0, nil
	}

	err = s.send()
	return s.lines, err
}

// SubmitEventuallyOnBackground eventually forks a child process that submits analytics to WeDeploy
// if the analytics reporting is enabled
func SubmitEventuallyOnBackground(conf *config.Config) (err error) {
	if !conf.EnableAnalytics {
		return nil
	}

	var s = &Sender{}
	if err = s.read(); err != nil {
		return err
	}

	if s.lines == 0 {
		return nil
	}

	return s.maybeSubmitOnBackground()
}

func isConnectedToIPv4() bool {
	var addrs, err = net.InterfaceAddrs()

	if err != nil {
		verbose.Debug(
			errwrap.Wrapf("Error trying to verify if it is connected to the Internet (ignoring): {{err}}", err))
		return true
	}

	for _, addr := range addrs {
		var ip = addr.(*net.IPNet)
		if !ip.IP.IsLoopback() && ip.IP.To4() != nil {
			return true
		}
	}

	return false
}

func (s *Sender) isOld() bool {
	var fileInfo, err = os.Stat(metricsPath)

	if err != nil {
		verbose.Debug(errwrap.Wrapf(
			"Error while trying to stat metrics file, ignoring and assuming it is old: {{err}}",
			err))
		return true
	}

	var timeout = fileInfo.ModTime().Add(6 * time.Hour)
	return time.Now().After(timeout)
}

func (s *Sender) maybeSubmitOnBackground() error {
	if !isConnectedToIPv4() && rand.Float32() > 0.5 {
		return nil
	}

	if s.lines <= 30 && rand.Float32() > 0.2 {
		return nil
	}

	if !s.isOld() && rand.Float32() > 0.8 {
		return nil
	}

	return s.submitOnBackground()
}

func (s *Sender) submitOnBackground() error {
	var cmd = exec.Command(os.Args[0], "metrics", "usage", "submit")

	if err := cmd.Start(); err != nil {
		return errwrap.Wrapf("Error trying to submit metrics on background: {{err}}", err)
	}

	verbose.Debug(fmt.Sprintf(
		`Started short-lived analytics reporting background process (PID %v)
Metrics stored in ~/.we_metrics are going to be submitted in batch to WeDeploy`,
		cmd.Process.Pid))
	return nil
}

func (s *Sender) read() (err error) {
	switch s.content, err = ioutil.ReadFile(metricsPath); {
	case os.IsNotExist(err):
		return nil
	case err != nil:
		return errwrap.Wrapf("Can not read metrics file: {{err}}", err)
	default:
		s.countLines()
		return nil
	}
}

func (s *Sender) countLines() {
	var sp = strings.SplitN(string(s.content), "\n", -1)
	s.lines = len(sp)

	if sp[s.lines-1] == "" {
		s.lines--
	}
}

func (s *Sender) send() (err error) {
	var request = wedeploy.URL(server)
	request.Body(bytes.NewReader(s.content))

	var ctx, cancelFunc = context.WithTimeout(context.Background(), MetricsSubmissionTimeout)
	request.SetContext(ctx)
	err = request.Post()
	verbosereq.Feedback(request)
	cancelFunc()

	return err
}

// Purge current metrics file
func Purge() (err error) {
	if err = os.Truncate(metricsPath, 0); err != nil {
		return errwrap.Wrapf(
			"Error trying to purge "+metricsPath+": please remove it by hand. Error: {{err}}",
			err)
	}

	return nil
}
