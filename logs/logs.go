package logs

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/colorwheel"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/verbose"
)

// Logs structure
type Logs struct {
	ServiceID  string `json:"serviceId"`
	ServiceUID string `json:"serviceUid"`
	DeployUID    string `json:"deployUid"`
	ProjectID    string `json:"projectId"`
	Level        int    `json:"level"`
	Message      string `json:"message"`
	Severity     string `json:"severity"`
	Timestamp    int64  `json:"timestamp"`
}

// Filter structure
type Filter struct {
	Project   string `json:"-"`
	Service string `json:"serviceId,omitempty"`
	Instance  string `json:"serviceUid,omitempty"`
	Level     int    `json:"level,omitempty"`
	Since     string `json:"start,omitempty"`
}

// Watcher structure
type Watcher struct {
	Filter          *Filter
	PoolingInterval time.Duration
	filterMutex     sync.Mutex
	end             bool
	endMutex        sync.Mutex
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
var outStreamMutex sync.Mutex

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
		err = errwrap.Wrapf(fmt.Sprintf(`Unknown log level "%v"`, severityOrLevel), err)
	}

	return i, err
}

// GetList logs
func GetList(ctx context.Context, filter *Filter) ([]Logs, error) {
	var list []Logs

	var params = []string{
		"/projects",
		url.QueryEscape(filter.Project),
	}

	if filter.Service != "" {
		params = append(params, "/services", url.QueryEscape(filter.Service))
	}

	params = append(params, "/logs")

	var req = apihelper.URL(ctx, params...)

	apihelper.Auth(req)

	if filter.Level != 0 {
		req.Param("level", fmt.Sprintf("%d", filter.Level))
	}

	if filter.Since != "" {
		req.Param("start", filter.Since)
	}

	var err = apihelper.Validate(req, req.Get())

	if err != nil {
		return list, errwrap.Wrapf("Can not list logs: {{err}}", err)
	}

	err = apihelper.DecodeJSON(req, &list)

	if err != nil {
		return list, errwrap.Wrapf("Can not decode logs JSON: {{err}}", err)
	}

	return filterInstanceInLogs(list, filter.Instance), nil
}

func filterInstanceInLogs(list []Logs, instance string) []Logs {
	if instance == "" {
		return list
	}

	var l = []Logs{}

	for _, il := range list {
		if strings.HasPrefix(il.ServiceUID, instance) {
			l = append(l, il)
		}
	}

	return l
}

// List logs
func List(ctx context.Context, filter *Filter) error {
	var list, err = GetList(ctx, filter)

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
		outStreamMutex.Lock()
		fmt.Fprintln(outStream, "")
		outStreamMutex.Unlock()
		watcher.Stop()
		done <- true
	}()

	<-done
}

// Start for Watcher
func (w *Watcher) Start() {
	go func() {
		for {
			w.endMutex.Lock()
			e := w.end
			w.endMutex.Unlock()

			if e {
				return
			}

			w.pool()
			time.Sleep(w.PoolingInterval)
		}
	}()
}

// Stop for Watcher
func (w *Watcher) Stop() {
	w.endMutex.Lock()
	w.end = true
	w.endMutex.Unlock()
}

func printList(list []Logs) {
	for _, log := range list {
		iw := instancesWheel.Get(log.ServiceUID)
		fd := color.Format(iw,
			log.ServiceID+"-"+
				log.ProjectID+"."+
				config.Context.RemoteAddress+
				"["+trim(log.ServiceUID, 7)+"]")
		outStreamMutex.Lock()
		fmt.Fprintf(outStream, "%v %v\n", fd, log.Message)
		outStreamMutex.Unlock()
	}
}

func (w *Watcher) pool() {
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	var list, err = GetList(ctx, w.Filter)

	cancel()

	if err != nil {
		fmt.Fprintf(errStream, "%v\n", errorhandling.Handle(err))
		return
	}

	printList(list)

	var length = len(list)

	if length == 0 {
		w.filterMutex.Lock()
		verbose.Debug("No new log since " + w.Filter.Since)
		w.filterMutex.Unlock()
		return
	}

	w.incSinceArgument(list)
}

func (w *Watcher) incSinceArgument(list []Logs) {
	var last = list[len(list)-1]
	var next = last.Timestamp + 1

	w.filterMutex.Lock()
	w.Filter.Since = fmt.Sprintf("%v", next)
	verbose.Debug("Next --since parameter value = " + w.Filter.Since)
	w.filterMutex.Unlock()
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
