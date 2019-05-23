package logs

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/wedeploy/cli/apihelper"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/colorwheel"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/errorhandler"
	"github.com/wedeploy/cli/verbose"
)

// Client for the services
type Client struct {
	*apihelper.Client
}

// New Client
func New(wectx config.Context) *Client {
	return &Client{
		apihelper.New(wectx),
	}
}

// Log structure
type Log struct {
	ServiceID    string `json:"serviceId"`
	ContainerUID string `json:"containerUid"`
	Build        bool   `json:"build"`
	DeployUID    string `json:"deployUid"`
	ProjectID    string `json:"projectId"`
	Level        int    `json:"level"`
	Message      string `json:"message"`
	Severity     string `json:"severity"`
	Timestamp    string `json:"timestamp"`
}

// Filter structure
type Filter struct {
	Project  string   `json:"-"`
	Services []string `json:"services,omitempty"`
	Instance string   `json:"containerUid,omitempty"`
	Level    int      `json:"level,omitempty"`
	Since    string   `json:"start,omitempty"`
}

// Watcher structure
type Watcher struct {
	Client          *Client
	PoolingInterval time.Duration

	Filter      *Filter
	filterMutex sync.Mutex

	ctx context.Context
}

// SeverityToLevel map
var SeverityToLevel = map[string]int{
	"critical": 2,
	"error":    3,
	"warning":  4,
	"info":     6,
	"debug":    7,
}

// PoolingInterval is the default time between retries.
var PoolingInterval = 5 * time.Second

var instancesWheel = colorwheel.New(color.TextPalette)

var errStream io.Writer = os.Stderr
var errStreamMutex sync.Mutex

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
		err = errwrap.Wrapf(fmt.Sprintf(`unknown log level "%v"`, severityOrLevel), err)
	}

	return i, err
}

// GetList logs
func (c *Client) GetList(ctx context.Context, f *Filter) ([]Log, error) {
	var list []Log

	var params = []string{
		"/projects",
		url.PathEscape(f.Project),
	}

	params = append(params, "/logs")

	var req = c.Client.URL(ctx, params...)

	c.Client.Auth(req)

	if len(f.Services) == 1 && f.Services[0] != "" {
		// avoid getting all logs unnecessarily
		// CAUTION: see filter function below: on changes here, update it.
		req.Param("serviceId", f.Services[0])
	}

	if f.Level != 0 {
		req.Param("level", fmt.Sprintf("%d", f.Level))
	}

	if f.Since != "" {
		req.Param("start", f.Since)
	}

	// it relies on wedeploy/data, which currently has a hard limit of 9999 results
	req.Param("limit", "9999")

	var err = apihelper.Validate(req, req.Get())

	if err != nil {
		return list, err
	}

	err = apihelper.DecodeJSON(req, &list)

	if err != nil {
		return list, errwrap.Wrapf("can't decode logs JSON: {{err}}", err)
	}

	return filter(list, f.Services, f.Instance), nil
}

func filter(list []Log, services []string, instance string) []Log {
	// CAUTION: see optimization call to ?serviceId=:serviceID above: on changes here, update it.
	if instance == "" && len(services) <= 1 {
		return list
	}

	var l = []Log{}

	for _, il := range list {
		if isService(il.ServiceID, services) && isContainer(il.ContainerUID, instance) {
			l = append(l, il)
		}
	}

	return l
}

func isContainer(s, prefix string) bool {
	if prefix == "" {
		return true
	}

	return strings.HasPrefix(s, prefix)
}

func isService(current string, ss []string) bool {
	for _, s := range ss {
		if current == s {
			return true
		}
	}

	return len(ss) == 0
}

// List logs
func (c *Client) List(ctx context.Context, filter *Filter) error {
	var list, err = c.GetList(ctx, filter)

	if err == nil {
		printList(list)
	}

	return err
}

// Watch logs. If no pooling interval is set it uses the default value.
func (w *Watcher) Watch(ctx context.Context, wectx config.Context) {
	w.ctx = ctx
	w.Client = New(wectx)

	if w.PoolingInterval == 0 {
		w.PoolingInterval = PoolingInterval
	}

	w.watch()
}

func (w *Watcher) watch() {
	w.pool()

	ticker := time.NewTicker(w.PoolingInterval)

	for {
		select {
		case <-w.ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			w.pool()
		}
	}
}

func addHeader(log Log) (m string) {
	m = log.ServiceID
	switch {
	case log.ServiceID == "":
		m += "[" + log.ProjectID + "]"
	case log.ContainerUID != "":
		m += "[" + trim(log.ContainerUID, 12) + "]"
	case log.Build:
		m += "[build]"
	}

	return m
}

func printList(list []Log) {
	for _, log := range list {
		iw := instancesWheel.Get(log.ProjectID + "-" + log.ContainerUID)
		fd := color.Format(iw, addHeader(log))
		ts := color.Format(color.FgWhite, getTimestamp(log.Timestamp))

		outStreamMutex.Lock()
		_, _ = fmt.Fprintf(outStream, "%v %v %v\n", ts, fd, log.Message)
		outStreamMutex.Unlock()
	}
}

func getTimestamp(timestamp string) string {
	i, err := strconv.ParseInt(timestamp, 10, 64)

	if err != nil {
		verbose.Debug("can't decode timestamp", err)
		return "unknown"
	}

	t := time.Unix(0, i)
	return t.Format("Jan 02 15:04:05.000")
}

func (w *Watcher) pool() {
	var ctx, cancel = context.WithTimeout(w.ctx, 10*time.Second)
	defer cancel()

	var list, err = w.Client.GetList(ctx, w.Filter)
	cancel()

	if err != nil && w.ctx.Err() == nil {
		errStreamMutex.Lock()
		defer errStreamMutex.Unlock()
		_, _ = fmt.Fprintf(errStream, "%v\n", errorhandler.Handle(err))
		return
	}

	if len(list) == 0 {
		w.filterMutex.Lock()
		defer w.filterMutex.Unlock()
		verbose.Debug("No new log since " + w.Filter.Since)
		return
	}

	printList(list)

	if err := w.incSinceArgument(list); err != nil {
		errStreamMutex.Lock()
		defer errStreamMutex.Unlock()
		_, _ = fmt.Fprintf(errStream, "%v\n", errorhandler.Handle(err))
		return
	}
}

func (w *Watcher) incSinceArgument(list []Log) error {
	var last = list[len(list)-1]

	var next, err = strconv.ParseUint(last.Timestamp, 10, 64)

	if err != nil {
		return errwrap.Wrapf("invalid timestamp value on log line: {{err}}", err)
	}

	next++

	w.filterMutex.Lock()
	defer w.filterMutex.Unlock()
	w.Filter.Since = fmt.Sprintf("%v", next)
	verbose.Debug("Next --since parameter value = " + w.Filter.Since)
	return nil
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
