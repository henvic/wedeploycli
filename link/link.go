package link

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"io"

	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/verbose"
)

// Machine structure
type Machine struct {
	ProjectID   string
	Links       []*Link
	ErrStream   io.Writer
	Errors      *Errors
	ErrorsMutex sync.Mutex
	dirMutex    sync.Mutex
	queue       sync.WaitGroup
	Watcher     *list.Watcher
	list        *list.List
	end         bool
}

// Link holds the information of container to be linked
type Link struct {
	Container     *containers.Container
	ContainerPath string
}

// ContainerError struct
type ContainerError struct {
	ContainerPath string
	Error         error
}

// Errors list
type Errors struct {
	List []ContainerError
}

// New Container link
func New(dir string) (*Link, error) {
	var l = &Link{
		ContainerPath: dir,
	}

	c, err := containers.Read(l.ContainerPath)
	l.Container = c

	if err != nil {
		return nil, err
	}

	verbose.Debug("Container ID " + c.ID + " for directory " + dir)

	return l, err
}

func (le Errors) Error() string {
	var msgs = []string{}

	for _, e := range le.List {
		msgs = append(msgs, fmt.Sprintf("%v: %v", e.ContainerPath, e.Error.Error()))
	}

	return fmt.Sprintf("Linking errors\n%v", strings.Join(msgs, "\n"))
}

// Setup the linking machine
func (m *Machine) Setup(list []string) error {
	m.Errors = &Errors{
		List: []ContainerError{},
	}

	for _, dir := range list {
		m.mount(dir)
	}

	return nil
}

// Watch changes due to linking
func (m *Machine) Watch() {
	var cs []string

	for _, l := range m.Links {
		cs = append(cs, l.Container.ID)
	}

	m.list = list.New(list.Filter{
		Project:    m.ProjectID,
		Containers: cs,
	})

	m.Watcher = list.NewWatcher(m.list)
	m.Watcher.StopCondition = m.isDone
	m.Watcher.Start()
}

func (m *Machine) isDone() bool {
	if !m.end {
		return false
	}

	if len(m.Errors.List) == 0 {
		if len(m.list.Projects) == 0 {
			return false
		}

		var projectWatched = m.list.Projects[0]

		for _, link := range m.Links {
			c, ok := projectWatched.Containers[link.Container.ID]

			if !ok || c.Health != "up" {
				return false
			}
		}
	} else {
		println("Killing linking watcher after linking errors (use \"we list\" to see what is up).")
	}

	return true
}

// Run links the containers of the list input
func (m *Machine) Run() {
	m.queue.Add(len(m.Links))
	m.linkAll()
	m.queue.Wait()
	m.end = true
}

func (m *Machine) linkAll() {
	for _, cl := range m.Links {
		// this might be in parallel (with a goroutine call)
		// but currently there is an issue on the backend
		// that makes things fail randomly
		// see https://github.com/wedeploy/cli/issues/170
		m.doLink(cl)
	}
}

func (m *Machine) doLink(cl *Link) {
	var err = m.link(cl)

	switch err {
	case nil:
		if m.ErrStream != nil {
			fmt.Fprintf(m.ErrStream, "Container %v linked.\n", cl.Container.ID)
		}
	default:
		m.logError(cl.ContainerPath, err)
	}

	m.queue.Done()
}

func (m *Machine) link(l *Link) error {
	m.dirMutex.Lock()
	var err = containers.Link(context.Background(), m.ProjectID, l.Container.ID, l.ContainerPath)
	m.dirMutex.Unlock()
	runtime.Gosched()

	return err
}

func (m *Machine) logError(dir string, err error) {
	m.ErrorsMutex.Lock()
	m.Errors.List = append(m.Errors.List, ContainerError{
		ContainerPath: dir,
		Error:         errorhandling.Handle(err),
	})

	m.ErrorsMutex.Unlock()
}

func (m *Machine) mount(dir string) {
	var l, err = New(dir)

	if err != nil {
		m.logError(dir, err)
		return
	}

	m.Links = append(m.Links, l)
}
