package logs

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/colorwheel"
	"github.com/wedeploy/cli/verbose"
)

// Logs structure
type Logs struct {
	ContainerID  string `json:"containerId"`
	ContainerUID string `json:"containerUid"`
	ProjectID    string `json:"projectId"`
	Level        int    `json:"level"`
	Message      string `json:"message"`
	Severity     string `json:"severity"`
	Timestamp    string `json:"timestamp"`
}

// Filter structure
type Filter struct {
	Project   string `json:"-"`
	Container string `json:"containerId,omitempty"`
	Instance  string `json:"containerUid,omitempty"`
	Level     int    `json:"level,omitempty"`
	Since     string `json:"start,omitempty"`
}

// Watcher structure
type Watcher struct {
	Filter          *Filter
	PoolingInterval time.Duration
	end             bool
}

// SeverityToLevel map
var SeverityToLevel = map[string]int{
	"critical": 2,
	"error":    3,
	"warning":  4,
	"info":     6,
	"debug":    7,
}

// PoolingInterval is the time between retries
var PoolingInterval = time.Second

var instancesWheel = colorwheel.New(color.TextPalette)

var errStream io.Writer = os.Stderr
var outStream io.Writer = os.Stdout

// GetLevel to get level from severity or itself
func GetLevel(severityOrLevel string) (int, error) {
	var level = SeverityToLevel[severityOrLevel]

	if level != 0 {
		return level, nil
	}

	if severityOrLevel == "" {
		return 0, nil
	}

	var i, err = strconv.Atoi(severityOrLevel)

	if err != nil {
		err = errwrap.Wrapf("Can't translate log severity param to level: {{err}}", err)
	}

	return i, err
}

// GetList logs
func GetList(filter *Filter) ([]Logs, error) {
	var list []Logs
	var req = apihelper.URL("/logs/" + filter.Project)

	apihelper.Auth(req)
	apihelper.ParamsFromJSON(req, filter)

	var err = apihelper.Validate(req, req.Get())

	if err != nil {
		return list, errwrap.Wrapf("Can't list logs: {{err}}", err)
	}

	err = apihelper.DecodeJSON(req, &list)

	if err != nil {
		return list, errwrap.Wrapf("Can't decode logs JSON: {{err}}", err)
	}

	return list, err
}

// List logs
func List(filter *Filter) error {
	var list, err = GetList(filter)

	if err == nil {
		printList(list)
	}

	return err
}

// Watch logs
func Watch(watcher *Watcher) {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	watcher.Start()

	go func() {
		<-sigs
		fmt.Fprintln(outStream, "")
		watcher.Stop()
		done <- true
	}()

	<-done
}

// Start for Watcher
func (w *Watcher) Start() {
	go func() {
	w:
		if !w.end {
			w.pool()
			time.Sleep(w.PoolingInterval)
		}

		if !w.end {
			goto w
		}
	}()
}

// Stop for Watcher
func (w *Watcher) Stop() {
	w.end = true
}

func printList(list []Logs) {
	for _, log := range list {
		iw := instancesWheel.Get(log.ContainerUID)
		fd := color.Format(iw, log.ContainerID+"."+log.ProjectID+".wedeploy.me["+trim(log.ContainerUID, 7)+"]")
		fmt.Fprintf(outStream, "%v %v\n", fd, log.Message)
	}
}

func (w *Watcher) pool() {
	var list, err = GetList(w.Filter)

	if err != nil {
		fmt.Fprintf(errStream, "%v\n", err)
		return
	}

	printList(list)

	var length = len(list)

	if length == 0 {
		verbose.Debug("No new log since " + w.Filter.Since)
		return
	}

	w.incSinceArgument(list)
}

func (w *Watcher) incSinceArgument(list []Logs) {
	var last = list[len(list)-1]
	var next, err = strconv.ParseInt(last.Timestamp, 10, 0)

	if err != nil {
		panic(err)
	}

	next++

	w.Filter.Since = fmt.Sprintf("%v", next)
	verbose.Debug("Next --since parameter value = " + w.Filter.Since)
}

// GetUnixTimestamp gets the Unix timestamp in seconds from a friendly string.
// Be aware that the dashboard is using ms, not s.
func GetUnixTimestamp(since string) (int64, error) {
	if num, err := strconv.ParseInt(since, 10, 0); err == nil {
		return num, err
	}

	var now = time.Now()

	since = strings.Replace(since, "min", "m", -1)

	var pds, err = time.ParseDuration(since)

	if err != nil {
		return 0, err
	}

	return now.Add(-pds).Unix(), err
}

func trim(s string, max int) string {
	runes := []rune(s)

	if len(runes) > max {
		return string(runes[:max])
	}

	return s
}
