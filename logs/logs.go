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

// LongPoolingPeriod is the time between retries
const LongPoolingPeriod = time.Second

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
	Level      int    `json:"level,omitempty"`
	InstanceID string `json:"instanceId,omitempty"`
	Since      string `json:"start,omitempty"`
}

// SeverityToLevel map
var SeverityToLevel = map[string]int{
	"critical": 2,
	"error":    3,
	"warning":  4,
	"info":     6,
	"debug":    7,
}

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
func GetList(filter Filter, paths ...string) []Logs {
	var list []Logs
	var req = apihelper.URL("/api/logs/" + strings.Join(paths, "/"))

	apihelper.Auth(req)
	apihelper.ParamsFromJSON(req, filter)

	apihelper.ValidateOrExit(req, req.Get())
	apihelper.DecodeJSON(req, &list)

	return list
}

// List logs
func List(filter Filter, paths ...string) {
	var list = GetList(filter, paths...)
	printList(list)
}

// Watch logs
func Watch(filter Filter, paths ...string) {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		done <- true
		fmt.Println()
		os.Exit(0)
	}()

	watch(filter, paths...)
	<-done
}

func printList(list []Logs) {
	for _, log := range list {
		fmt.Fprintln(outStream, log.Message)
	}
}

func watch(filter Filter, paths ...string) {
	for {
		var list = GetList(filter, paths...)

		printList(list)

		time.Sleep(LongPoolingPeriod)

		var length = len(list)

		if length == 0 {
			verbose.Debug("No new log since " + filter.Since)
			continue
		}

		last := list[length-1]
		next, err := strconv.ParseInt(last.Timestamp, 10, 0)

		if err != nil {
			panic(err)
		}

		next++

		filter.Since = fmt.Sprintf("%v", next)
		verbose.Debug("Next --since parameter value = " + filter.Since)
	}
}
