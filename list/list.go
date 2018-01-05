package list

import (
	"context"
	"errors"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/projects"
	"github.com/wedeploy/cli/services"
)

// Filter parameters for the list command
type Filter struct {
	Project  string
	Services []string
}

// List services object
type List struct {
	Detailed bool

	Filter          Filter
	PoolingInterval time.Duration

	Projects   []projects.Project
	lastError  error
	updated    chan bool
	watchMutex sync.RWMutex

	SelectNumber  bool
	ProjectHeader func(projectID string) string
	ServiceHeader func(serviceID, projectID string) string
	ProjectFooter func(projectID string) string
	ServiceFooter func(serviceID, projectID string) string

	livew     *uilive.Writer
	outStream io.Writer

	projectsClient *projects.Client
	servicesClient *services.Client

	w *formatter.TabWriter

	retry int

	wectx     config.Context
	ctx       context.Context
	stop      context.CancelFunc
	selectors []Selection
}

// Selection of a list
type Selection struct {
	Project string
	Service string
}

// New creates a list using the values of a passed Filter
func New(filter Filter) *List {
	var l = &List{
		Filter:          filter,
		PoolingInterval: time.Second,
		updated:         make(chan bool, 1),
	}

	return l
}

func (l *List) prepare(ctx context.Context, wectx config.Context) {
	l.ctx, l.stop = context.WithCancel(ctx)
	l.wectx = wectx

	l.projectsClient = projects.New(l.wectx)
	l.servicesClient = services.New(l.wectx)

	l.livew = uilive.New()
	l.outStream = l.livew
	l.w = formatter.NewTabWriter(l.outStream)
}

// Start for the list
func (l *List) Start(ctx context.Context, wectx config.Context) {
	l.prepare(ctx, wectx)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	go l.update()
	go l.watch()
	go l.watchKiller(sigs)

	<-l.ctx.Done()
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
}

// Once runs the list only once
func (l *List) Once(ctx context.Context, wectx config.Context) error {
	l.PoolingInterval = time.Minute
	l.prepare(ctx, wectx)
	l.updateHandler()
	l.w.Init(l.outStream)
	l.watchHandler()

	l.watchMutex.RLock()
	var le = l.lastError
	l.watchMutex.RUnlock()
	return le
}

// GetSelection of service or project
func (l *List) GetSelection(option string) (Selection, error) {
	var num, err = strconv.Atoi(option)

	if err != nil {
		return Selection{
			Project: option,
		}, nil
	}

	var sel = l.selectors

	if len(sel) < num {
		return Selection{}, errors.New("invalid selection")
	}

	var s = sel[num-1]
	return s, nil
}

func isContextError(err error) bool {
	if err == nil {
		return false
	}

	if strings.Contains(err.Error(), context.DeadlineExceeded.Error()) ||
		strings.Contains(err.Error(), context.Canceled.Error()) {
		return true
	}

	return false
}
