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

	"github.com/launchpad-project/cli/apihelper"
	"github.com/launchpad-project/cli/verbose"
)

// Logs structure
type Logs struct {
	AppName    string `json:"appName"`
	InstanceID string `json:"instanceId"`
	Level      int    `json:"level"`
	Message    string `json:"message"`
	PodName    string `json:"podName"`
	Severity   string `json:"severity"`
	Timestamp  string `json:"timestamp"`
}

// Filter structure
type Filter struct {
	Level int    `json:"level,omitempty"`
	Since string `json:"start,omitempty"`
}

// Watcher structure
type Watcher struct {
	Filter          *Filter
	Paths           []string
	PoolingInterval time.Duration
	ticker          *time.Ticker
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

	return strconv.Atoi(severityOrLevel)
}

// GetList logs
func GetList(filter *Filter, paths ...string) []Logs {
	var list []Logs
	var req = apihelper.URL("/api/logs/" + strings.Join(paths, "/"))

	apihelper.Auth(req)
	apihelper.ParamsFromJSON(req, filter)

	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSON(req, &list)

	return list
}

// List logs
func List(filter *Filter, paths ...string) {
	var list = GetList(filter, paths...)
	printList(list)
}

// Watch logs
func Watch(watcher *Watcher) {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	watcher.Start()

	go func() {
		<-sigs
		done <- true
		fmt.Fprintln(outStream, "")
		watcher.Stop()
	}()

	<-done
}

// Start for Watcher
func (w *Watcher) Start() {
	w.ticker = time.NewTicker(w.PoolingInterval)

	go func() {
		w.pool()
		for range w.ticker.C {
			w.pool()
		}
	}()
}

// Stop for Watcher
func (w *Watcher) Stop() {
	w.ticker.Stop()
}

func printList(list []Logs) {
	for _, log := range list {
		fmt.Fprintln(outStream, log.Message)
	}
}

func (w *Watcher) pool() {
	var list = GetList(w.Filter, w.Paths...)

	printList(list)

	var length = len(list)

	if length == 0 {
		verbose.Debug("No new log since " + w.Filter.Since)
		return
	}

	last := list[length-1]
	next, err := strconv.ParseInt(last.Timestamp, 10, 0)

	if err != nil {
		panic(err)
	}

	next++

	w.Filter.Since = fmt.Sprintf("%v", next)
	verbose.Debug("Next --since parameter value = " + w.Filter.Since)
}
