package link

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"

	"github.com/wedeploy/cli/containers"
	"github.com/wedeploy/cli/errorhandling"
	"github.com/wedeploy/cli/list"
	"github.com/wedeploy/cli/verbose"
)

// Machine structure
type Machine struct {
	ProjectID   string
	Links       []*Link
	Errors      *Errors
	FErrStream  io.Writer
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

	return fmt.Sprintf("List of errors (format is container path: error)\n%v",
		strings.Join(msgs, "\n"))
}

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
	m.Watcher.StopCondition = m.linkedContainersUp
	m.Watcher.Start()
}

func (m *Machine) linkedContainersUp() bool {
	if !m.end || len(m.list.Projects) == 0 {
		return false
	}

	var projectWatched = m.list.Projects[0]

	for _, link := range m.Links {
		c, ok := projectWatched.Containers[link.Container.ID]

		if !ok || c.Health != "up" {
			return false
		}
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
		go m.doLink(cl)
	}
}

func (m *Machine) doLink(cl *Link) {
	var err = m.link(cl)

	if err != nil {
		m.logError(cl.ContainerPath, err)
	}

	m.queue.Done()
}

func (m *Machine) link(l *Link) error {
	m.dirMutex.Lock()
	var err = containers.Link(m.ProjectID,
		l.ContainerPath,
		l.Container)
	m.dirMutex.Unlock()
	runtime.Gosched()

	return err
}

func (m *Machine) logError(dir string, err error) {
	m.ErrorsMutex.Lock()
	m.Errors.List = append(m.Errors.List, ContainerError{
		ContainerPath: dir,
		Error:         err,
	})

	if m.FErrStream != nil {
		fmt.Fprintf(m.FErrStream, "%v error: %v\n", dir, errorhandling.Handle(err))
	}

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
