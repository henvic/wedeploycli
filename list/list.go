package list

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/henvic/uilive"
	"github.com/henvic/wedeploycli/config"
	"github.com/henvic/wedeploycli/formatter"
	"github.com/henvic/wedeploycli/projects"
	"github.com/henvic/wedeploycli/services"
)

// Filter parameters for the list command
type Filter struct {
	Project  string
	Services []string

	HideServices bool
}

// Pattern for detailed listing
type Pattern uint

const (
	// Instances info
	Instances Pattern = 1 << iota
	// CPU info
	CPU
	// Memory info
	Memory
	// CreatedAt info
	CreatedAt
	// Detailed prints all details
	Detailed = Instances | CPU | Memory | CreatedAt
)

var details = []Pattern{Instances, CPU, Memory, CreatedAt}

// List services object
type List struct {
	Details Pattern

	Filter          Filter
	PoolingInterval time.Duration

	Projects   []projects.Project
	lastError  error
	updated    chan struct{}
	watchMutex sync.RWMutex

	AllowCreateProjectOnPrompt bool
	SelectNumber               bool

	livew     *uilive.Writer
	outStream io.Writer

	projectsClient *projects.Client
	servicesClient *services.Client

	w *formatter.TabWriter

	once bool

	retry int

	wectx     config.Context
	ctx       context.Context
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
		updated:         make(chan struct{}, 1),
	}

	return l
}

func (l *List) prepare(ctx context.Context, wectx config.Context) {
	l.ctx = ctx
	l.wectx = wectx

	l.projectsClient = projects.New(l.wectx)
	l.servicesClient = services.New(l.wectx)

	l.livew = uilive.New()
	l.outStream = l.livew
	l.w = formatter.NewTabWriter(l.outStream)
}

// Watch for the list
func (l *List) Watch(ctx context.Context, wectx config.Context) {
	l.prepare(ctx, wectx)

	go l.update()
	l.watch()
}

// Once runs the list only once
func (l *List) Once(ctx context.Context, wectx config.Context) error {
	l.once = true

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
