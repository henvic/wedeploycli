package listinstances

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/henvic/uilive"
	"github.com/wedeploy/cli/config"
	"github.com/wedeploy/cli/formatter"
	"github.com/wedeploy/cli/services"
)

// List services object
type List struct {
	Project         string
	Service         string
	PoolingInterval time.Duration

	Instances  []services.Instance
	lastError  error
	updated    chan struct{}
	watchMutex sync.RWMutex

	SelectNumber bool

	livew     *uilive.Writer
	outStream io.Writer

	servicesClient *services.Client

	w *formatter.TabWriter

	once bool

	retry int

	wectx     config.Context
	ctx       context.Context
	selectors []string
}

// New creates a list using the values of a passed Filter
func New(projectID, serviceID string) *List {
	return &List{
		Project:         projectID,
		Service:         serviceID,
		PoolingInterval: time.Second,
		updated:         make(chan struct{}, 1),
	}
}

func (l *List) fetch() ([]services.Instance, error) {
	ctx, cancel := context.WithTimeout(l.ctx, 30*time.Second)
	defer cancel()
	return l.servicesClient.Instances(ctx, l.Project, l.Service)
}

func (l *List) prepare(ctx context.Context, wectx config.Context) {
	l.ctx = ctx
	l.wectx = wectx

	l.servicesClient = services.New(l.wectx)

	l.livew = uilive.New()
	l.outStream = l.livew
	l.w = formatter.NewTabWriter(l.outStream)
}

// Watch list instances
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
