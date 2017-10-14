package list

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/color"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/projects"
)

// Filter parameters for the list command
type Filter struct {
	Project  string
	Services []string
}

// List services object
type List struct {
	Detailed           bool
	Filter             Filter
	PoolingInterval    time.Duration
	Projects           []projects.Project
	HandleRequestError func(error) string
	StopCondition      (func() bool)
	Teardown           func(l *List)
	end                chan bool
	livew              *uilive.Writer
	outStream          io.Writer
	w                  *formatter.TabWriter
	retry              int
	killed             bool
	killLock           sync.Mutex
	wectx              config.Context
}

// New creates a list using the values of a passed Filter
func New(filter Filter) *List {
	var l = &List{
		Filter:          filter,
		PoolingInterval: time.Second,
	}

	l.HandleRequestError = l.handleRequestError
	return l
}

// Printf list
func (l *List) Printf(format string, a ...interface{}) {
	fmt.Fprintf(l.w, format, a...)
}

func (l *List) printList() {
	l.w.Init(l.outStream)
	var ps, err = l.fetch()
	l.Projects = ps

	if err != nil {
		l.Printf("%v", l.HandleRequestError(err))
		return
	}

	l.retry = 0
	l.printProjects()
	_ = l.w.Flush()
}

func (l *List) handleRequestError(err error) string {
	l.retry++
	return fmt.Sprintf("%v %v #%d\n", color.Format(color.BgHiRed, color.FgRed, "!"), errorhandling.Handle(err), l.retry)
}

// Start for the list
func (l *List) Start(wectx config.Context) {
	l.wectx = wectx
	sigs := make(chan os.Signal, 1)
	l.end = make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	l.livew = uilive.New()
	l.outStream = l.livew
	l.w = formatter.NewTabWriter(l.outStream)

	go l.watch()
	go l.watchKill(sigs)

	<-l.end
	l.finish()
}

func (l *List) watchKill(sigs chan os.Signal) {
	<-sigs
	l.Stop()
	l.killLock.Lock()
	l.killed = true
	l.killLock.Unlock()
}

func (l *List) finish() {
	if l.Teardown != nil {
		l.Teardown(l)
	}
}

func (l *List) watch() {
p:
	l.printList()
	l.Flush()

	if l.StopCondition != nil && l.StopCondition() {
		l.Stop()
		return
	}

	time.Sleep(l.PoolingInterval)
	goto p
}

// Flush list
func (l *List) Flush() {
	_ = l.livew.Flush()
}

// Stop for the list
func (l *List) Stop() {
	l.end <- true
}
